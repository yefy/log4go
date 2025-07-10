//go:build !log4_debug
// +build !log4_debug

package log4

const IsLog4Debug = false

func log4Debug(format string, args ...interface{}) {
}
