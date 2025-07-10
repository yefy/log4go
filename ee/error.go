package ee

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

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
	funcName = runtime.FuncForPC(pc).Name()
	funcName = GetLastStrPart(funcName, ".")
	if !ok {
		file = "???"
		line = 0
	}
	//addrStr := fmt.Sprintf("%s:%d %s", file, line, funcName)
	addrStr := fmt.Sprintf("%s:%d|%s", file, line, funcName)

	if err != nil {
		if len(err_str) <= 0 {
			return errors.New(fmt.Sprintf("[ERROR] %s @@@ %v", addrStr, err))
		}
		return errors.New(fmt.Sprintf("[ERROR] %s =》 %s @@@ %v", err_str, addrStr, err))
	}
	if len(err_str) <= 0 {
		return errors.New(fmt.Sprintf("[ERROR] %s", addrStr))
	}
	return errors.New(fmt.Sprintf("[ERROR] %s =》 %s", err_str, addrStr))
}

func GetLastStrPart(s string, substr string) string {
	lastDot := strings.LastIndex(s, substr)
	if lastDot == -1 {
		return ""
	}
	return s[lastDot+len(substr):]
}
