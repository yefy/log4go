package log4

import (
	"github.com/yefy/log4go/ee"
	"os"
	"unsafe"
)

func ModTime(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, ee.New(err, "os.Stat")
	}

	modTime := info.ModTime()
	return modTime.Unix(), nil
}

func SliceByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToSliceByte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
