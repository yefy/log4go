//go:build linux
// +build linux

package efile

import (
	"os"
)

func OpenFileWithShareDelete(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
}
