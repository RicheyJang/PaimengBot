package utils

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func PathJoin(paths ...string) string {
	return filepath.ToSlash(filepath.Join(paths...))
}
