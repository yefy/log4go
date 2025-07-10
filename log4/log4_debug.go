//go:build log4_debug
// +build log4_debug

package log4

import "fmt"

const IsLog4Debug = true

func log4Debug(format string, args ...interface{}) {
	fmt.Printf(">>>>>>>>>>>>>>>>>>>>log4Debug|"+format+"\n", args...)
}
