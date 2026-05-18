package ee

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

const isShortPath = true

var openStackInfoToErrorLog bool

func OpenStackInfoToErrorLog(b bool) {
	openStackInfoToErrorLog = b
}

type Error struct {
	gCode int
	code  int
	msg   string
	stack []uintptr
	cause error
}

func NewErr(err error) error {
	if err == nil {
		return nil
	}
	return DoNew(err, 2, "")
}

func New(err error, format string, a ...any) error {
	return DoNew(err, 2, format, a...)
}

func NewCode(code int, err error, format string, a ...any) error {
	return DoNew(err, 2, format, a...).CodeSet(code)
}

func NewCopyCode(err error, format string, a ...any) error {
	e := DoNew(err, 2, format, a...)
	if ee, ok := err.(*Error); ok {
		e.CodeSet(ee.Code())
	}
	return e
}

func DoNew(err error, skip int, format string, a ...any) *Error {
	msg := buildMessage(skip+1, format, a...)

	if e, ok := err.(*Error); ok {
		return &Error{
			gCode: e.gCode,
			code:  0,
			msg:   msg,
			stack: e.stack,
			cause: err,
		}
	}

	return &Error{
		gCode: 0,
		code:  0,
		msg:   msg,
		stack: callers(skip + 1),
		cause: err,
	}
}

func buildMessage(skip int, format string, a ...any) string {
	var errStr string

	if len(a) > 0 {
		errStr = fmt.Sprintf(format, a...)
	} else {
		errStr = format
	}

	pc, file, line, ok := runtime.Caller(skip)

	var funcName string

	if !ok {
		file = "???"
		line = 0
		funcName = "???"
	} else {
		if isShortPath {
			file = TrimPathN(file, 3)
		}

		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = GetLastStrPart(fn.Name(), ".")
		}
	}

	return fmt.Sprintf(
		" %s:%d @%s emsg(%s)",
		file,
		line,
		funcName,
		errStr,
	)
}

func callers(skip int) []uintptr {
	pcs := make([]uintptr, 64)

	n := runtime.Callers(skip, pcs)

	return pcs[:n]
}

func (e *Error) IgnorePrint() bool {
	return e.gCode < 0
}

func (e *Error) GCode() int {
	return e.gCode
}

func (e *Error) GCodeSet(code int) *Error {
	e.gCode = code
	return e
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) CodeSet(code int) *Error {
	e.code = code
	return e
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Stack() []uintptr {
	return e.stack
}

func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil {
		_, _ = io.WriteString(s, "<nil>")
		return
	}

	switch verb {

	case 's':
		_, _ = io.WriteString(s, e.msg)
		return

	case 'q':
		fmt.Fprintf(s, "%q", e.msg)
		return

	case 'v':
		// %+#v
		if s.Flag('+') && s.Flag('#') {
			e.formatFull(s, true)
			return
		}

		// %+v
		if s.Flag('+') {
			e.formatFull(s, openStackInfoToErrorLog)
			return
		}

		// %v
		e.formatFull(s, openStackInfoToErrorLog)
		return
	}
}

func (e *Error) formatFull(w io.Writer, isPrintStack bool) {
	index := 0

	fmt.Fprintf(w, "\n")
	cur := error(e)
	for cur != nil {
		if ee, ok := cur.(*Error); ok {
			fmt.Fprintf(w, "    %d.%s\n", index, ee.msg)
			cur = ee.cause
		} else {
			fmt.Fprintf(w, "    %d.%v\n", index, cur)
			break
		}

		index++
	}

	if isPrintStack {
		// stack
		if len(e.stack) > 0 {

			fmt.Fprintf(w, "\nStack:\n")

			frames := runtime.CallersFrames(e.stack)

			for {
				frame, more := frames.Next()
				file := frame.File
				if isShortPath {
					file = TrimPathN(frame.File, 4)
				}

				fmt.Fprintf(
					w,
					"%s\n\t%s:%d\n",
					frame.Function,
					file,
					frame.Line,
				)

				if !more {
					break
				}
			}
		}
	}
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
	last := strings.LastIndex(s, substr)

	if last == -1 {
		return s
	}

	return s[last+len(substr):]
}
