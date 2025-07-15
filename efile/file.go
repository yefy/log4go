package efile

import (
	"log4go/ee"
	"os"
	"path/filepath"
)

func EnsureLogDirExists(logPath string) error {
	dir := filepath.Dir(logPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return ee.New(nil, "failed to create directory %s: %w", dir, mkErr)
		}
	}

	return nil
}
