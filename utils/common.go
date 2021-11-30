package utils

import (
	"errors"
	"runtime"
	"sync"

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
