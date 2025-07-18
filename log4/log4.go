package log4

import (
	"fmt"
	"github.com/yefy/log4go/ee"
	"github.com/yefy/log4go/efile"
	"io/ioutil"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

var newlineRe = regexp.MustCompile(`\r?\n`)

// / END_OF_LINE
var endOfLine = "<<EOL>>"

var defaultRootTarget = "root"

// var defaultDiscardTarget = "discard_root"
var defaultDiscardTarget = defaultRootTarget

var GCtx = NewWaitGroupContext()

var GLog4 atomic.Pointer[Log4]

func init() {
	log4 := NewLog4("")
	GCtx.Add(1)
	GLog4.Store(log4)
}

func NewLog4(path string) *Log4 {
	RecordCountStatAdd(Log4StartCount)
	log4 := &Log4{path: path, appenderMap: make(map[string]Log4Appender), context: NewWaitGroupContext()}
	log4.TargetMap.Store(defaultRootTarget, NewLog4Target(defaultRootTarget))
	log4.TargetMap.Store(defaultDiscardTarget, NewLog4Target(defaultDiscardTarget))
	return log4
}

type Log4 struct {
	path string
	//string *Log4Target
	TargetMap   sync.Map
	appenderMap map[string]Log4Appender
	RefreshRate int64
	deferFuncs  []func()
	context     *WaitGroupContext
	IsClose     bool
}

func (log4 *Log4) Target(targetName string) *Log4Target {
	log4TargetI, ok := log4.TargetMap.Load(targetName)
	if !ok {
		return nil
	}
	return log4TargetI.(*Log4Target)
}

func (log4 *Log4) Run(log4Config *Log4Config) error {
	err := log4Config.Check()
	if err != nil {
		return ee.New(err, "log4Config.Check")
	}
	log4.RefreshRate = log4Config.RefreshRate

	for name, v := range log4Config.Appenders {
		if v.Kind == KindConsole {
			appender := NewLog4ConsoleAppender(name, &v)
			log4.appenderMap[name] = appender
		} else if v.Kind == KindFile {
			file, err := efile.OpenFileWithShareDelete(v.Path)
			if err != nil {
				return ee.New(err, "open path:%v ", v.Path)
			}

			appender := NewLog4FileAppender(name, &v, file)
			log4.appenderMap[name] = appender
		} else {
			return ee.New(err, "not find kind:%v", v.Kind)
		}
	}

	for _, appender := range log4.appenderMap {
		appender.Run()
	}

	createTargetFunc := func(targetName string, logger *Log4ConfigLogger, rootTarget *Log4Target) (*Log4Target, error) {
		target := NewLog4Target(targetName)
		target.Name = targetName
		target.Level = LevelNameToLevelDef(logger.Level)
		target.Logger = logger
		target.RootTarget = rootTarget
		for _, name := range logger.Appenders {
			appender, ok := log4.appenderMap[name]
			if !ok {
				return nil, ee.New(err, "not find appender:%v", name)
			}
			target.appenders = append(target.appenders, appender)
		}
		return target, nil
	}

	rootTarget, err := createTargetFunc(defaultRootTarget, &log4Config.Root, nil)
	if err != nil {
		return ee.New(err, "")
	}
	log4.TargetMap.Store(defaultRootTarget, rootTarget)

	for name, logger := range log4Config.Loggers {
		rootTarget, err := createTargetFunc(name, &logger, rootTarget)
		if err != nil {
			return ee.New(err, "")
		}
		log4.TargetMap.Store(name, rootTarget)
	}

	log4.context.Add(1)
	go func() {
		RecordCountStatAdd(ReInitFileStartCount)
		ticker := time.NewTicker(time.Duration(log4.RefreshRate) * time.Second)
		defer func() {
			ticker.Stop()
			RecordCountStatAdd(ReInitFileEndCount)
			recordCountStatPrint()
			log4.context.Done()
		}()

		done := log4.context.Ctx.Done()

		lastModTime, _ := ModTime(log4.path)
		for {
			select {
			case <-done:
				log4Debug("reInitFile done")
				return
			case <-ticker.C:
				modTime, _ := ModTime(log4.path)
				if lastModTime == modTime {
					continue
				}
				lastModTime = modTime
				log4Debug("reInitFile")
				InitFile(log4.path)
			}
		}
	}()

	return nil
}

func (log4 *Log4) Flush() {
	for _, appender := range log4.appenderMap {
		appender.Flush()
	}
}

func (log4 *Log4) Close(isWait bool) {
	RecordCountStatAdd(Log4EndCount)
	for _, appender := range log4.appenderMap {
		appender.Close(isWait)
	}

	log4.context.Quit(isWait)

	for _, deferFunc := range log4.deferFuncs {
		deferFunc()
	}
}

func NewLog4Target(name string) *Log4Target {
	log4Target := &Log4Target{
		Name:  name,
		Level: ERROR,
	}
	return log4Target
}

type Log4Target struct {
	Name       string
	Level      Level
	Logger     *Log4ConfigLogger
	RootTarget *Log4Target
	appenders  []Log4Appender
}

func (log4Target *Log4Target) GetLevel() Level {
	return log4Target.Level
}

func (log4Target *Log4Target) Critical(format string, args ...interface{}) {
	log4Target.log(3, CRITICAL, format, args...)
}

func (log4Target *Log4Target) Error(format string, args ...interface{}) {
	log4Target.log(3, ERROR, format, args...)
}

func (log4Target *Log4Target) Warn(format string, args ...interface{}) {
	log4Target.log(3, WARNING, format, args...)
}

func (log4Target *Log4Target) Info(format string, args ...interface{}) {
	log4Target.log(3, INFO, format, args...)
}
func (log4Target *Log4Target) Debug(format string, args ...interface{}) {
	log4Target.log(3, DEBUG, format, args...)
}

func (log4Target *Log4Target) Trace(format string, args ...interface{}) {
	log4Target.log(3, TRACE, format, args...)
}

func (log4Target *Log4Target) Fine(format string, args ...interface{}) {
	log4Target.log(3, FINE, format, args...)
}

func (log4Target *Log4Target) rootCritical(format string, args ...interface{}) {
	log4Target.log(4, CRITICAL, format, args...)
}

func (log4Target *Log4Target) rootError(format string, args ...interface{}) {
	log4Target.log(4, ERROR, format, args...)
}

func (log4Target *Log4Target) rootWarn(format string, args ...interface{}) {
	log4Target.log(4, WARNING, format, args...)
}

func (log4Target *Log4Target) rootInfo(format string, args ...interface{}) {
	log4Target.log(4, INFO, format, args...)
}
func (log4Target *Log4Target) rootDebug(format string, args ...interface{}) {
	log4Target.log(4, DEBUG, format, args...)
}

func (log4Target *Log4Target) rootTrace(format string, args ...interface{}) {
	log4Target.log(4, TRACE, format, args...)
}

func (log4Target *Log4Target) rootFine(format string, args ...interface{}) {
	log4Target.log(4, FINE, format, args...)
}

func (log4Target *Log4Target) log(skip int, level Level, format string, args ...interface{}) {
	if level < log4Target.Level {
		return
	}

	rec := log4Target.GetRecord(skip, level, format, args...)
	defer rec.Put()

	log4Target.WriteRecord(rec)

	if log4Target.Logger != nil && log4Target.Logger.Additive && log4Target.RootTarget != nil {
		log4Target.RootTarget.WriteRecord(rec)
	}
}

func (log4Target *Log4Target) WriteRecord(rec *Log4Record) {
	for _, appender := range log4Target.appenders {
		appender.LogRecord(rec.Clone())
	}
}

func (log4Target *Log4Target) GetRecord(skip int, level Level, format string, args ...interface{}) *Log4Record {
	// Determine caller func
	var funcName string
	var pc uintptr
	var file string
	var line int
	var ok bool
	pc, file, line, ok = runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
		funcName = "???"
	} else {
		file = ee.TrimPathN(file, 3)
		funcName = runtime.FuncForPC(pc).Name()
		funcName = ee.GetLastStrPart(funcName, ".")
	}
	src := fmt.Sprintf("%s:%d@%s", file, line, funcName)

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	if !log4Target.Logger.Multiline {
		msg = newlineRe.ReplaceAllString(msg, endOfLine)
	}

	// Make the log record
	rec := NewLog4Record()
	rec.Target = log4Target.Name
	rec.Level = LevelToLevelFileName(level)
	rec.Created = time.Now()
	rec.CreatedUtc = time.Now().UTC()
	rec.Source = src
	rec.Message = msg

	return rec
}

func GetLevel() Level {
	return Target(defaultRootTarget).GetLevel()
}

func Critical(format string, args ...interface{}) {
	Target(defaultRootTarget).rootCritical(format, args...)
}

func Error(format string, args ...interface{}) {
	Target(defaultRootTarget).rootError(format, args...)
}

func Warn(format string, args ...interface{}) {
	Target(defaultRootTarget).rootWarn(format, args...)
}

func Info(format string, args ...interface{}) {
	Target(defaultRootTarget).rootInfo(format, args...)
}
func Debug(format string, args ...interface{}) {
	Target(defaultRootTarget).rootDebug(format, args...)
}

func Trace(format string, args ...interface{}) {
	Target(defaultRootTarget).rootTrace(format, args...)
}

func Fine(format string, args ...interface{}) {
	Target(defaultRootTarget).rootFine(format, args...)
}

func Target(targetName string) *Log4Target {
	log4 := (*Log4)(GLog4.Load())
	target := log4.Target(targetName)
	if target == nil {
		target = log4.Target(defaultDiscardTarget)
	}
	return target
}

func Flush() {
	log4 := (*Log4)(GLog4.Load())
	log4.Flush()
}

func Close(isWait bool) {
	log4 := (*Log4)(GLog4.Load())
	if log4.IsClose {
		return
	}
	log4.IsClose = true
	if !isWait {
		log4.Close(isWait)
		GCtx.Done()
		return
	}

	log4.Close(isWait)
	GCtx.Done()
	GCtx.Wait()
}

func Reopen() {
	log4 := (*Log4)(GLog4.Load())
	if log4.IsClose {
		log4Debug("err:Reopen => log4 is close")
		return
	}
	log4Debug("Reopen")
	InitFile(log4.path)
}

func InitFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ee.New(err, "ioutil.ReadFile path:%v", path)
	}
	log4Config := &Log4Config{}
	err = yaml.Unmarshal(data, log4Config)
	if err != nil {
		return ee.New(err, "yaml.Unmarshal path:%v", path)
	}
	log4Debug("log4Config:%+v", log4Config)

	log4 := NewLog4(path)
	err = log4.Run(log4Config)
	if err != nil {
		return ee.New(err, "log4.Run path:%v", path)
	}

	GCtx.Add(1)
	log4 = GLog4.Swap(log4)
	go func() {
		defer GCtx.Done()
		time.Sleep(time.Second * 2)
		log4.Close(true)
	}()

	return nil
}
