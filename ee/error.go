package ee

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

// / END_OF_LINE
var endOfLine = "<<EOL>>"

const (
	displayFormat = 0
	debugFormat   = 1
	fullFormat    = 2
)

func newError(msgs []string, stack []uintptr, cause error) *Error {
	return &Error{msgs: msgs, stack: stack, cause: cause}
}

type Error struct {
	msgs  []string
	stack []uintptr
	cause error
}

func (obj *Error) formatStack(w io.Writer) {
	frames := runtime.CallersFrames(obj.stack)

	for {
		frame, more := frames.Next()

		fmt.Fprintf(
			w,
			"%s\n\t%s:%d\n",
			frame.Function,
			frame.File,
			frame.Line,
		)

		if !more {
			break
		}
	}
}

func callers(skip int) []uintptr {
	pcs := make([]uintptr, 32)

	n := runtime.Callers(skip, pcs)

	return pcs[:n]
}

func (obj *Error) Unwrap() error {
	return obj.cause
}

func (obj *Error) Msgs() []string {
	return obj.msgs
}

func (obj *Error) Stack() []uintptr {
	return obj.stack
}

func (obj *Error) Error() string {
	value := fmt.Sprintf("debug.error%s", endOfLine)
	for i := 0; i < len(obj.msgs); i++ {
		msg := obj.msgs[len(obj.msgs)-1-i]
		value = fmt.Sprintf("%s    %d: %s%s", value, i, msg, endOfLine)
	}

	return value
}

func (obj *Error) Format(s fmt.State, verb rune) {
	if obj == nil {
		fmt.Fprintf(s, "<nil>")
		return
	}

	format := displayFormat
	switch verb {
	case 'v':
		if s.Flag('+') {
			format = debugFormat
		}

		if s.Flag('+') && s.Flag('#') {
			format = fullFormat
		}
	case 's':

	case 'q':
	}

	if format == fullFormat {
		fmt.Fprintf(s, "debug.error\n")
		for i := 0; i < len(obj.msgs); i++ {
			msg := obj.msgs[len(obj.msgs)-1-i]
			fmt.Fprintf(s, "    %d: %s\n", i, msg)
		}
		obj.formatStack(s)
	} else if format == debugFormat {
		fmt.Fprintf(s, "debug.error\n")
		for i := 0; i < len(obj.msgs); i++ {
			msg := obj.msgs[len(obj.msgs)-1-i]
			fmt.Fprintf(s, "    %d: %s\n", i, msg)
		}
	} else {
		fmt.Fprintf(s, "debug.error%s", endOfLine)
		for i := 0; i < len(obj.msgs); i++ {
			msg := obj.msgs[len(obj.msgs)-1-i]
			fmt.Fprintf(s, "    %d: %s%s", i, msg, endOfLine)
		}
	}
}

func New(err error, format string, a ...any) error {
	return DoNew(err, 2, format, a...)
}

func DoNew(err error, skip int, format string, a ...any) error {
	var errStr string
	if len(a) > 0 {
		errStr = fmt.Sprintf(format, a...)
	} else {
		errStr = format
	}

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
		file = TrimPathN(file, 3)
		funcName = runtime.FuncForPC(pc).Name()
		funcName = GetLastStrPart(funcName, ".")
	}
	errStr = fmt.Sprintf("%s:%d@%s emsg(%s)", file, line, funcName, errStr)

	var cause error
	msgs, stack := func() ([]string, []uintptr) {
		if err != nil {
			log4Err, ok := err.(*Error)
			if ok {
				cause = log4Err.cause
				return log4Err.Msgs(), log4Err.Stack()
			} else {
				cause = err
				return []string{err.Error()}, nil
			}
		}
		return nil, nil
	}()

	if len(stack) <= 0 {
		stack = callers(2)
	}

	msgs = append(msgs, errStr)

	return newError(msgs, stack, cause)
}

func TrimPathN(file string, keep int) string {
	slashPath := filepath.ToSlash(file)
	parts := strings.Split(slashPath, "/")
	if len(parts) <= keep {
		return slashPath
	}
	return strings.Join(parts[len(parts)-keep:], "/")
}

func GetLastStrPart(s string, substr string) string {
	lastDot := strings.LastIndex(s, substr)
	if lastDot == -1 {
		return ""
	}
	return s[lastDot+len(substr):]
}
