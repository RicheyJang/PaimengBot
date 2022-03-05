package push

import "github.com/tidwall/gjson"

func intersectionWithJsonArray(a []int64, b []gjson.Result, bKey string) []int64 {
	m := make(map[int64]bool)
	for _, v := range a {
		m[v] = true
	}
	var ret []int64
	for _, obj := range b {
		v := obj.Get(bKey).Int()
		if m[v] {
			ret = append(ret, v)
		}
	}
	return ret
}
