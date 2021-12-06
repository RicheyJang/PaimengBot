package utils

import (
	"encoding/json"
	"errors"
	"regexp"
	"runtime"
	"sync"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
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

// MergeStringSlices 合并多个字符串切片并去重、去除空字符串
func MergeStringSlices(slices ...[]string) (res []string) {
	mp := FormSetByStrings(slices...)
	for s, _ := range mp {
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
