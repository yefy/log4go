package ee

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// / END_OF_LINE
var endOfLine = "<<EOL>>"

func New(err error, format string, a ...any) error {
	return DoNew(err, 2, format, a...)
}

func DoNew(err error, skip int, format string, a ...any) error {
	err_str := ""
	if len(a) <= 0 {
		err_str = fmt.Sprintf("%s", format)
	} else {
		err_str = fmt.Sprintf(format, a...)
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
	addrStr := fmt.Sprintf("%s:%d@%s", file, line, funcName)

	if err != nil {
		if len(err_str) <= 0 {
			return errors.New(fmt.Sprintf("[%s emsg()]%s%v", addrStr, endOfLine, err))
		}
		return errors.New(fmt.Sprintf("[%s emsg(%s)]%s%v", addrStr, err_str, endOfLine, err))
	}
	if len(err_str) <= 0 {
		return errors.New(fmt.Sprintf("[%s emsg()]", addrStr))
	}
	return errors.New(fmt.Sprintf("[%s emsg(%s)]", addrStr, err_str))
}
func TrimPathN(file string, keep int) string {
	slashPath := filepath.ToSlash(file) // 转为统一斜杠
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
