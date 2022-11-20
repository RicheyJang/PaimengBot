package utils

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"regexp"
	"runtime"
	"sync"
	"unicode"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func GoAndWait(handlers ...func() error) (err error) {
	var wg sync.WaitGroup
	var once sync.Once
	for _, f := range handlers {
		wg.Add(1)
		go func(handler func() error) {
			defer func() {
				if e := recover(); e != nil {
					buf := make([]byte, 1024)
					buf = buf[:runtime.Stack(buf, false)]
					log.Errorf("[GoAndWait PANIC]%v\n%s\n", e, buf)
					once.Do(func() {
						err = errors.New("panic found in call handlers")
					})
				}
				wg.Done()
			}()
			if e := handler(); e != nil {
				once.Do(func() {
					err = e
				})
			}
		}(f)
	}
	wg.Wait()
	return err
}

// JsonString 将任意内容转换为Json字符串
func JsonString(v interface{}) string {
	res, _ := json.Marshal(v)
	return string(res)
}

// StringLimit 限制字符串长度，若超出limit，返回前limit个码点+"..."
func StringLimit(s string, limit int) string {
	runeSlice := []rune(s)
	if len(runeSlice) <= limit {
		return s
	}
	return string(runeSlice[:limit]) + "..."
}

// MergeStringSlices 合并多个字符串切片并去重、去除空字符串
func MergeStringSlices(slices ...[]string) (res []string) {
	mp := FormSetByStrings(slices...)
	for s := range mp {
		if len(s) == 0 {
			continue
		}
		res = append(res, s)
	}
	return
}

// FormSetByStrings 将字符串切片形成Set
func FormSetByStrings(slices ...[]string) map[string]struct{} {
	mp := make(map[string]struct{})
	for _, slice := range slices {
		for _, s := range slice {
			mp[s] = struct{}{}
		}
	}
	return mp
}

// StringSliceContain 字符串切片中是否含有指定字符串
func StringSliceContain(slices []string, substr string) bool {
	for _, str := range slices {
		if str == substr {
			return true
		}
	}
	return false
}

// DeleteStringInSlice 删除字符串切片中的str元素，并去重
func DeleteStringInSlice(slice []string, str ...string) []string {
	slice = MergeStringSlices(slice)
	str = MergeStringSlices(str)
	for _, s := range str {
		for i, now := range slice {
			if now == s {
				slice = append(slice[:i], slice[i+1:]...)
				break
			}
		}
	}
	return slice
}

var letterReg = regexp.MustCompile(`^[A-Za-z]+$`)

// IsLetter 字符串是否为纯字母
func IsLetter(s string) bool {
	return letterReg.MatchString(s)
}

var numberReg = regexp.MustCompile(`^\d+$`)

// IsNumber 字符串是否为纯数字
func IsNumber(s string) bool {
	return numberReg.MatchString(s)
}

// StringRealLength 计算字符串的真实长度
func StringRealLength(s string) int {
	return utf8.RuneCountInString(s)
}

// SplitOnSpace 按文字、空格、文字...分隔字符串
func SplitOnSpace(x string) []string {
	var result []string
	pi := 0
	ps := false
	for i, c := range x {
		s := unicode.IsSpace(c)
		if s != ps && i > 0 {
			result = append(result, x[pi:i])
			pi = i
		}
		ps = s
	}
	result = append(result, x[pi:])
	return result
}

// UInt32ToBytes uint32转换为bytes
func UInt32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

// BytesToUInt32 bytes转换为uint32
func BytesToUInt32(b []byte) uint32 {
	for len(b) < 4 {
		b = append(b, byte(0))
	}
	return binary.LittleEndian.Uint32(b)
}

//将int切片转换为int64切片
func IntSlice2int64Slice(slice []int) (newslice []int64) {
	for i := 0; i < len(slice); i++ {
		newslice = append(newslice, int64(slice[i]))
	}
	return
}

//将string切片转换为int64切片
func StringSlice2int64Slice(slice []string) (newslice []int64) {
	for i := 0; i < len(slice); i++ {
		newslice = append(newslice, cast.ToInt64(slice[i]))
	}
	return
}
