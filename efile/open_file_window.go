//go:build windows
// +build windows

package efile

import (
	"os"
	"syscall"
)

func OpenFileWithShareDelete(path string) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	//access := uint32(syscall.GENERIC_WRITE)
	access := uint32(syscall.FILE_APPEND_DATA)

	shareMode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE | syscall.FILE_SHARE_DELETE)

	creationDisposition := uint32(syscall.OPEN_ALWAYS)

	flags := uint32(syscall.FILE_ATTRIBUTE_NORMAL)

	handle, err := syscall.CreateFile(
		p,
		access,
		shareMode,
		nil,
		creationDisposition,
		flags,
		0,
	)
	if err != nil {
		return nil, err
	}

	//_, err = syscall.SetFilePointer(handle, 0, nil, syscall.FILE_END)
	//if err != nil {
	//	syscall.CloseHandle(handle)
	//	return nil, err
	//}

	return os.NewFile(uintptr(handle), path), nil
}
