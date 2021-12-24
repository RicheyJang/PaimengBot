package genshin_draw

import (
	"encoding/binary"
)

func getKVNum(key string) uint32 {
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return 0
	}
	return BytesToUInt32(v)
}

func putKVNum(key string, value uint32) error {
	return proxy.GetLevelDB().Put([]byte(key), UInt32ToBytes(value), nil)
}

func UInt32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func BytesToUInt32(b []byte) uint32 {
	for len(b) < 4 {
		b = append(b, byte(0))
	}
	return binary.LittleEndian.Uint32(b)
}
