package log4

import (
	"fmt"
	"io"
	"os"

	"github.com/yefy/log4go/ee"
)

const (
	defaultBufSize   = 4096
)

//让用户控制尝试几次
var flushMaxRetries = 3
func FlushBuffToFileMaxRetries(n int)  {
	flushMaxRetries = n
}

type Log4Writer struct {
	buf []byte
	n   int
	wr  io.Writer
}

func NewLog4WriterSize(w io.Writer, size int) *Log4Writer {
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
	written, err := b.writeAll(b.buf[0:b.n])
	if written > 0 {
		copy(b.buf, b.buf[written:b.n])
		b.n -= written
	}
	return err
}

func (b *Log4Writer) writeAll(buf []byte) (int, error) {
	bufSize := len(buf)
	if bufSize <= 0 {
		return 0, nil
	}

	total := 0
	var lastErr error
	retries := 0

	for total < bufSize {
		n, err := b.wr.Write(buf[total:])
		if n > 0 {
			total += n
			retries = 0
			if total == bufSize {
				return total, nil
			}
			continue
		}
		if err != nil {
			lastErr = err
		}
		retries++
		if retries >= flushMaxRetries {
			break
		}
	}

	if total == bufSize {
		return total, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("write stalled at %d/%d bytes", total, bufSize)
	}

	logWriteError("flush failed: wrote %d/%d bytes, err: %v", total, bufSize, lastErr)
	return total, ee.New(lastErr, "err:Flush")
}

func (b *Log4Writer) Write(s []byte) (int, error) {
	sLen := len(s)
	if sLen >= b.Size() {
		if err := b.Flush(); err != nil {
			return 0, err
		}
		n, err := b.writeAll(s)
		return n, err
	}
	if sLen > b.Available() {
		if err := b.Flush(); err != nil {
			return 0, err
		}
	}

	n := copy(b.buf[b.n:], s)
	if n != sLen {
		fmt.Fprintf(os.Stderr, "log4: err:copy => n:%v != sLen:%v\n", n, sLen)
	}
	b.n += n
	return sLen, nil
}
