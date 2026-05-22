package log4

import (
	"bytes"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var isOpenMsgPool bool = true

func SetMsgPoolFlag(isOpen bool) {
	isOpenMsgPool = isOpen
}

var msgPoolsBlock int = 1024
var msgPools [1024]sync.Pool

func init() {
	for i := range msgPools {
		i := i
		func(i int) {
			msgPools[i].New = func() interface{} {
				msg := NewMsg((i+1)*msgPoolsBlock, &msgPools[i])
				return msg
			}
		}(i)
	}
}

func GetMsg(size int) *Msg {
	// 0 1024-1   0
	// 1024 1024*2-1 1
	index := (size + msgPoolsBlock - 1) / msgPoolsBlock
	if !isOpenMsgPool || index >= len(msgPools) {
		msg := NewMsg((index+1)*msgPoolsBlock, nil)
		msg.Init()
		return msg
	} else {
		for {
			msg := msgPools[index].Get().(*Msg)
			if msg.RefCount.Load() > 0 {
				log4Debug("log_tcp_msg GetMsg RefCount > 0, Stack:%s", string(debug.Stack()))
				panic("log_tcp_msg: Msg reference counter went negative (double Put? or use-after-free?)")
				continue
			}
			msg.Init()
			return msg
		}
	}
}

type Msg struct {
	RefCount atomic.Int32
	*bytes.Buffer
	Cap  int
	Pool *sync.Pool
}

func NewMsg(cap int, Pool *sync.Pool) *Msg {
	msg := &Msg{
		Buffer:   bytes.NewBuffer(make([]byte, 0, cap)),
		Pool:     Pool,
		Cap:      cap,
		RefCount: atomic.Int32{},
	}
	//msg.RefCount.Add(1)
	return msg
}

func (msg *Msg) Init() {
	msg.Reset()
	msg.RefCount.Store(1)
}

func (msg *Msg) GetData() []byte {
	if msg.RefCount.Load() <= 0 {
		log4Debug("log_tcp_msg GetData RefCount <= 0, msg.Cap:%v, Stack:%s", msg.Cap, string(debug.Stack()))
		panic("log_tcp_msg: Msg reference counter went negative (double Put? or use-after-free?)")
	}
	return msg.Bytes()
}

func (msg *Msg) Clone() *Msg {
	if msg.RefCount.Load() <= 0 {
		log4Debug("log_tcp_msg Clone RefCount <= 0, msg.Cap:%v, Stack:%s", msg.Cap, string(debug.Stack()))
		panic("log_tcp_msg: Msg reference counter went negative (double Put? or use-after-free?)")
	}
	msg.RefCount.Add(1)
	return msg
}

func (msg *Msg) Put() {
	RefCount := msg.RefCount.Add(-1)
	if RefCount == 0 {
		if msg.Pool != nil {
			msg.Pool.Put(msg)
		}
	} else if RefCount < 0 {
		log4Debug("log_tcp_msg RefCount < 0, Stack:%s", string(debug.Stack()))
		panic("log_tcp_msg: Msg reference counter went negative (double Put? or use-after-free?)")
	}
}
