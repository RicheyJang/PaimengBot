package utils

import (
	"io/ioutil"
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

// PathSize 获取指定路径文件大小，若路径为文件夹，可能会导致效率较低
func PathSize(path string) uint64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if fi.IsDir() {
		return getDirSizeSlow(path)
	}
	return uint64(fi.Size())
}

// 获取文件夹占用空间大小（效率较低）
func getDirSizeSlow(dirPath string) uint64 {
	dirSize := uint64(0)
	files, e := ioutil.ReadDir(dirPath)
	if e != nil {
		return 0
	}
	for _, f := range files {
		if f.IsDir() {
			dirSize += getDirSizeSlow(dirPath + "/" + f.Name())
		} else {
			dirSize += uint64(f.Size())
		}
	}
	return dirSize
}

// MakeDir 创建文件夹并返回文件夹绝对路径
func MakeDir(path string) (string, error) {
	return MakeDirWithMode(path, os.ModePerm)
}

// MakeDirWithMode 依据文件夹权限创建文件夹并返回文件夹绝对路径
func MakeDirWithMode(path string, perm os.FileMode) (string, error) {
	if DirExists(path) {
		return filepath.Abs(path)
	}
	err := os.MkdirAll(path, perm)
	if err == nil {
		return filepath.Abs(path)
	}
	return path, err
}

// RemovePath 删除指定路径文件或目录
func RemovePath(path string) error {
	if PathExists(path) {
		return os.RemoveAll(path)
	}
	return nil
}

// PathJoin 文件路径合并（并标准化）
func PathJoin(paths ...string) string {
	return filepath.ToSlash(filepath.Join(paths...))
}
