package log4

import (
	"fmt"
	"github.com/yefy/log4go/ee"
	"io"
)

const (
	defaultBufSize = 4096
)

type Log4Writer struct {
	buf []byte
	n   int
	wr  io.Writer
}

func NewLog4WriterSize(w io.Writer, size int) *Log4Writer {
	// Is it already a Writer?
	b, ok := w.(*Log4Writer)
	if ok && len(b.buf) >= size {
		return b
	}
	if size <= 0 {
		size = defaultBufSize
	}
	return &Log4Writer{
		buf: make([]byte, size),
		wr:  w,
	}
}

func NewLog4Writer(w io.Writer) *Log4Writer {
	return NewLog4WriterSize(w, defaultBufSize)
}

func (b *Log4Writer) Size() int { return len(b.buf) }

func (b *Log4Writer) Available() int { return len(b.buf) - b.n }

func (b *Log4Writer) Buffered() int { return b.n }

func (b *Log4Writer) Reset() { b.n = 0 }

func (b *Log4Writer) Flush() error {
	if b.n <= 0 {
		return nil
	}
	err := b.FlushBuf(b.buf[0:b.n])
	b.Reset()
	return err
}

func (b *Log4Writer) FlushBuf(buf []byte) error {
	bufSize := len(buf)
	if bufSize <= 0 {
		return nil
	}
	for i := 0; i < 3; i++ {
		n, err := b.wr.Write(buf)
		if err == nil && n == bufSize {
			return nil
		}

		if err != nil {
			fmt.Printf("err:Flush => index:%v, err:%v\n", i, err)
		} else {
			fmt.Printf("err:Flush => index:%v, n:%v != bufSize:%v\n", i, n, bufSize)
		}
	}

	return ee.New(nil, "err:Flush")
}

func (b *Log4Writer) Write(s []byte) (int, error) {
	sLen := len(s)
	if sLen >= b.Size() {
		b.Flush()
		b.FlushBuf(s)
		return sLen, nil
	} else if sLen > b.Available() {
		b.Flush()
	}

	n := copy(b.buf[b.n:], s)
	if n != sLen {
		fmt.Printf("err:copy => n:%v != sLen:%v", n, sLen)
	}
	b.n += n
	return sLen, nil
}

func (b *Log4Writer) WriteString(s string) (int, error) {
	return b.Write(StringToSliceByte(s))
}
