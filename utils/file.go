package utils

import (
	"os"
	"path/filepath"
)

// PathExists 判断路径（包括文件与文件夹）是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// FileExists 判断文件是否存在（若为文件夹，仍返回false）
func FileExists(path string) bool {
	fs, err := os.Stat(path)
	return (err == nil || os.IsExist(err)) && !fs.IsDir()
}

// DirExists 判断文件夹是否存在（若为文件，仍返回false）
func DirExists(path string) bool {
	fs, err := os.Stat(path)
	return (err == nil || os.IsExist(err)) && fs.IsDir()
}

// MakeDir 创建文件夹并返回文件夹绝对路径
func MakeDir(path string) (string, error) {
	if DirExists(path) {
		return filepath.Abs(path)
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err == nil {
		return filepath.Abs(path)
	}
	return path, err
}

func RemovePath(path string) error {
	if PathExists(path) {
		return os.RemoveAll(path)
	}
	return nil
}

func PathJoin(paths ...string) string {
	return filepath.ToSlash(filepath.Join(paths...))
}
