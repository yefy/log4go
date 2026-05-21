package log4

import (
	"context"
	"fmt"
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/efile"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const runThreadCount = "thread_count_start"
const stopThreadCount = "thread_count_stop"

const ReInitFileStartCount = "reinitfile_start"
const ReInitFileEndCount = "reinitfile_end"

const Log4StartCount = "log4_start"
const Log4EndCount = "log4_end"

var recordCountStat sync.Map

func recordCountStatGet(name string) *atomic.Int32 {
	countI, ok := recordCountStat.Load(name)
	if !ok {
		countI, _ = recordCountStat.LoadOrStore(name, &atomic.Int32{})
	}
	count := countI.(*atomic.Int32)
	return count
}

func RecordCountStatAdd(name string) {
	recordCountStatAdd(name)
}

func recordCountStatAdd(name string) {
	if !IsLog4Debug {
		return
	}
	count := recordCountStatGet(name)
	count.Add(1)
}

func recordCountStatPrint() {
	if !IsLog4Debug {
		return
	}

	type Value struct {
		Name  string
		Count *atomic.Int32
	}

	values := make([]*Value, 0, 100)
	recordCountStat.Range(func(key, value any) bool {
		name := key.(string)
		count := value.(*atomic.Int32)
		values = append(values, &Value{
			Name:  name,
			Count: count,
		})
		return true
	})
	sort.Slice(values, func(i, j int) bool {
		return values[i].Name < values[j].Name
	})

	log4Debug("-------------------------recordCountStat start")
	for _, value := range values {
		log4Debug("recordCountStat name:%v, count:%v", value.Name, value.Count.Load())
	}
	log4Debug("-------------------------recordCountStat end")
}

func init() {
	if IsLog4Debug {
		go func() {
			for {
				time.Sleep(time.Second * 3)
				recordCountStatPrint()
			}
		}()
	}
}

type Log4Appender interface {
	Name() string
	LogRecord(rec *Log4Record)
	Run()
	Flush()
	Close(isWait bool)
	WaitClose()
	BufferReset()
	BufferReopen() error
	BufferWrite(msg *Msg) error
	BufferFlush() error
	BufferSize() int
	BufferClose() error
}

func BufferWriteAndDropRec(log Log4Appender, context *Log4AppenderContext, rec *Log4Record, formatCache *formatCacheType) error {
	defer rec.Put()
	msg := FormatLogRecord(context.Appender.Pattern, context.IsUtc, rec, formatCache)
	if msg != nil {
		defer msg.Put()
		recordCountStatAdd(context.nameWrite)
		if err := log.BufferWrite(msg.Clone()); err != nil {
			logWriteError("appender %s write failed: %v", context.name, err)
			return err
		}
	}
	return nil
}

func drainRecChanPending(log Log4Appender, context *Log4AppenderContext, formatCache *formatCacheType) {
	for {
		select {
		case rec, ok := <-context.recChan:
			recordCountStatAdd(context.nameIn)
			if !ok {
				return
			}
			recordCountStatAdd(context.nameValid)
			BufferWriteAndDropRec(log, context, rec, formatCache)
		default:
			return
		}
	}
}

func drainRecChanAll(log Log4Appender, context *Log4AppenderContext, formatCache *formatCacheType) {
	for {
		drained := false
		for {
			select {
			case rec, ok := <-context.recChan:
				recordCountStatAdd(context.nameIn)
				if !ok {
					return
				}
				recordCountStatAdd(context.nameValid)
				BufferWriteAndDropRec(log, context, rec, formatCache)
				drained = true
			default:
				goto nextPass
			}
		}
	nextPass:
		if !drained {
			return
		}
	}
}

func BufferFlush(log Log4Appender, context *Log4AppenderContext, formatCache *formatCacheType) {
	drainRecChanPending(log, context, formatCache)
	if err := log.BufferFlush(); err != nil {
		logWriteError("appender %s buffer flush failed: %v", context.name, err)
	}
}

func Run(log Log4Appender, context *Log4AppenderContext) {
	context.context.Add(1)
	go func() {
		recordCountStatAdd(runThreadCount)
		ticker := time.NewTicker(1 * time.Second)
		formatCache := formatCacheType{}
		defer func() {
			if r := recover(); r != nil {
				logWriteError("appender %s panic: %v\n%s", context.name, r, debug.Stack())
			}
			ticker.Stop()
			drainRecChanAll(log, context, &formatCache)
			if err := log.BufferFlush(); err != nil {
				logWriteError("appender %s final buffer flush failed: %v", context.name, err)
			}
			if err := log.BufferClose(); err != nil {
				logWriteError("appender %s buffer close failed: %v", context.name, err)
			}
			recordCountStatAdd(stopThreadCount)
			recordCountStatPrint()
			context.context.Done()
		}()

		done := context.context.Ctx.Done()
		writeCount := 0
		lastWriteCount := writeCount
		var writeErr error
		for {
			select {
			case rec, ok := <-context.recChan:
				recordCountStatAdd(context.nameIn)
				if !ok {
					return
				}
				recordCountStatAdd(context.nameValid)
				//有错误, 又有新的数据, 需要清空上次数据, 写新的数据, 不能堵塞, 影响业务
				if writeErr != nil {
					log.BufferReset()
				}
				writeErr = BufferWriteAndDropRec(log, context, rec, &formatCache)
				writeCount += 1
			case <-done:
				log4Debug("record %v done", context.name)
				drainRecChanAll(log, context, &formatCache)
				if err := log.BufferFlush(); err != nil {
					logWriteError("appender %s shutdown buffer flush failed: %v", context.name, err)
				}
				return
			case <-context.flushChan:
				BufferFlush(log, context, &formatCache)
			case <-ticker.C:
				//如果没有新的数据, 尝试flush
				if lastWriteCount == writeCount && log.BufferSize() > 0 {
					if err := log.BufferFlush(); err != nil {
						logWriteError("appender %s periodic buffer flush failed: %v", context.name, err)
					}
					log4Debug("log.BufferFlush:lastWriteCount == writeCount")
				} else {
					//写入错误了, 有新的数据, 定时reopen
					if writeErr != nil {
						err := log.BufferReopen()
						if err != nil {
							logWriteError("appender %s periodic buffer flush failed: %v", context.name, err)
						} else {
							writeErr = nil
						}
					}
				}
				lastWriteCount = writeCount
			}
		}
	}()
}

type Log4AppenderContext struct {
	name            string
	nameIn          string
	nameValid       string
	nameWrite       string
	nameFlush       string
	nameClose       string
	nameRecordStart string
	nameRecordEnd   string
	Appender        *Log4ConfigAppender
	recChan         chan *Log4Record
	context         *WaitGroupContext
	flushChan       chan bool
	IsUtc           bool
	stopped         atomic.Bool
}

func (context *Log4AppenderContext) stopAccepting() {
	context.stopped.Store(true)
}

func (context *Log4AppenderContext) sendRecord(rec *Log4Record) {
	if context.stopped.Load() {
		rec.Put()
		return
	}
	select {
	case context.recChan <- rec:
	case <-context.context.Ctx.Done():
		rec.Put()
	}
}

func (context *Log4AppenderContext) closeAppender(isWait bool) {
	context.stopAccepting()
	context.context.Quit(isWait)
}

func (context *Log4AppenderContext) waitClose() {
	context.context.Wait()
}

func NewLog4FileAppender(name string, Appender *Log4ConfigAppender, file *os.File) *Log4FileAppender {
	isUtc := strings.Contains(Appender.Pattern, FORMAT_TIME_UTC)
	writer := NewLog4Writer(file)
	return &Log4FileAppender{
		Context: Log4AppenderContext{
			name:            name,
			nameIn:          name + "_in",
			nameValid:       name + "_valid",
			nameWrite:       name + "_write",
			nameFlush:       name + "_flush",
			nameClose:       name + "_close",
			nameRecordStart: name + "_record_stat",
			nameRecordEnd:   name + "_record_end",
			Appender:        Appender,
			recChan:         make(chan *Log4Record, 1024),
			context:         NewWaitGroupContext(),
			flushChan:       make(chan bool, 10),
			IsUtc:           isUtc,
		},
		File:   file,
		writer: writer,
	}
}

type Log4FileAppender struct {
	Context Log4AppenderContext
	File    *os.File

	writer *Log4Writer
}

func (log *Log4FileAppender) Name() string {
	return log.Context.name
}

func (log *Log4FileAppender) LogRecord(rec *Log4Record) {
	recordCountStatAdd(log.Context.nameRecordStart)
	log.Context.sendRecord(rec)
	recordCountStatAdd(log.Context.nameRecordEnd)
}

func (log *Log4FileAppender) Run() {
	Run(log, &log.Context)
}

func (log *Log4FileAppender) BufferWrite(msg *Msg) error {
	if msg == nil {
		return nil
	}
	defer msg.Put()

	_, err := log.writer.Write(msg.Bytes())
	if err != nil {
		//写入失败缓存无法容纳当前msg的数据了, 保留当前日志, 没啥用, 要
		log.writer.Reset()
		log4Debug("log.writer.WriteString err:%v", err)
		return err
	}
	return nil
}


func (log *Log4FileAppender) BufferReset() {
	log.writer.Reset()
}

func (log *Log4FileAppender) BufferReopen() error {
	file, err := efile.OpenFileWithShareDelete(log.Context.Appender.Path)
	if err != nil {
		return ee.New(err, "open path:%v ", log.Context.Appender.Path)
	}
	writer := NewLog4Writer(file)

	log.File.Close()
	log.File = file
	log.writer = writer
	return nil
}

func (log *Log4FileAppender) BufferFlush() error {
	if log.BufferSize() > 0 {
		recordCountStatAdd(log.Context.nameFlush)
		err := log.writer.Flush()
		if err != nil {
			log4Debug("log.writer.Flush err:%v", err)
			return err
		}
	}
	return nil
}

func (log *Log4FileAppender) BufferClose() error {
	err := log.BufferFlush()
	if err != nil {
		return err
	}
	recordCountStatAdd(log.Context.nameClose)
	err = log.File.Close()
	if err != nil {
		log4Debug("log.File.Close err:%v", err)
		return err
	}
	return nil
}

func (log *Log4FileAppender) BufferSize() int {
	return log.writer.Buffered()
}

func (log *Log4FileAppender) Flush() {
	log.Context.flushChan <- true
}

func (log *Log4FileAppender) Close(isWait bool) {
	log.Context.closeAppender(isWait)
}

func (log *Log4FileAppender) WaitClose() {
	log.Context.waitClose()
}

func NewLog4ConsoleAppender(name string, Appender *Log4ConfigAppender) *Log4ConsoleAppender {
	isUtc := strings.Contains(Appender.Pattern, FORMAT_TIME_UTC)
	return &Log4ConsoleAppender{
		Context: Log4AppenderContext{
			name:            name,
			nameIn:          name + "_in",
			nameValid:       name + "_valid",
			nameWrite:       name + "_write",
			nameFlush:       name + "_flush",
			nameClose:       name + "_close",
			nameRecordStart: name + "_record_stat",
			nameRecordEnd:   name + "_record_end",
			Appender:        Appender,
			recChan:         make(chan *Log4Record, 1024),
			context:         NewWaitGroupContext(),
			flushChan:       make(chan bool, 10),
			IsUtc:           isUtc,
		},
	}
}

type Log4ConsoleAppender struct {
	Context Log4AppenderContext
}

func (log *Log4ConsoleAppender) Name() string {
	return log.Context.name
}

func (log *Log4ConsoleAppender) LogRecord(rec *Log4Record) {
	recordCountStatAdd(log.Context.nameRecordStart)
	log.Context.sendRecord(rec)
	recordCountStatAdd(log.Context.nameRecordEnd)
}

func (log *Log4ConsoleAppender) Run() {
	Run(log, &log.Context)
}

func (log *Log4ConsoleAppender) BufferReset() {
}

func (log *Log4ConsoleAppender) BufferReopen() error {
	return nil
}

func (log *Log4ConsoleAppender) BufferWrite(msg *Msg) error {
	if msg == nil {
		return nil
	}
	defer msg.Put()
	fmt.Printf("%v", SliceByteToString(msg.Bytes()))
	return nil
}

func (log *Log4ConsoleAppender) BufferFlush() error {
	//recordCountStatAdd(log.Context.nameFlush)
	return nil
}

func (log *Log4ConsoleAppender) BufferClose() error {
	recordCountStatAdd(log.Context.nameClose)
	return nil
}

func (log *Log4ConsoleAppender) BufferSize() int {
	return 0
}

func (log *Log4ConsoleAppender) Flush() {
	log.Context.flushChan <- true
}

func (log *Log4ConsoleAppender) Close(isWait bool) {
	log.Context.closeAppender(isWait)
}

func (log *Log4ConsoleAppender) WaitClose() {
	log.Context.waitClose()
}

func NewLog4Record() *Log4Record {
	rec := Log4RecordPool.Get().(*Log4Record)
	rec.Init()
	return rec
}

type Log4Record struct {
	RefCount   atomic.Int32
	Pool       *sync.Pool
	Target     string
	Level      string
	Created    time.Time
	CreatedUtc time.Time
	Source     string
	Message    string
}

func (record *Log4Record) GetCreateTime(isUtc bool) time.Time {
	if isUtc {
		return record.CreatedUtc
	} else {
		return record.Created
	}
}

func (record *Log4Record) Init() {
	if record.RefCount.Load() > 0 {
		log4Debug("err: Log4Record Init RefCount > 0, Stack:%s", string(debug.Stack()))
	}
	record.Pool = &Log4RecordPool
	record.RefCount.Store(1)
}

func (record *Log4Record) Clone() *Log4Record {
	if record.RefCount.Load() <= 0 {
		log4Debug("err: Log4Record Clone RefCount <= 0, Stack:%s", string(debug.Stack()))
	}
	record.RefCount.Add(1)
	return record
}

func (record *Log4Record) Put() {
	RefCount := record.RefCount.Add(-1)
	if RefCount == 0 {
		if record.Pool != nil {
			record.Pool.Put(record)
		}
	} else if RefCount < 0 {
		log4Debug("err: Log4Record Put RefCount < 0, Stack:%s", string(debug.Stack()))
	}
}

var Log4RecordPool sync.Pool

func init() {
	Log4RecordPool.New = func() any {
		return &Log4Record{}
	}
}

func NewWaitGroupContext() *WaitGroupContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &WaitGroupContext{
		WaitGroup: &sync.WaitGroup{},
		Ctx:       ctx,
		Cancel:    cancel,
		IsQuit:    false,
	}
}

type WaitGroupContext struct {
	WaitGroup *sync.WaitGroup
	Ctx       context.Context
	Cancel    context.CancelFunc
	IsQuit    bool
}

func (w *WaitGroupContext) Quit(isWait bool) {
	if w.IsQuit {
		return
	}
	w.IsQuit = true
	w.Cancel()
	if isWait {
		w.Wait()
	}
}

func (w *WaitGroupContext) Wait() {
	w.WaitGroup.Wait()
}

func (w *WaitGroupContext) Add(delta int) {
	w.WaitGroup.Add(delta)
}

func (w *WaitGroupContext) Done() {
	w.WaitGroup.Done()
}
