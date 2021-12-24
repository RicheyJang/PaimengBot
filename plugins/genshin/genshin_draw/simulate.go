package genshin_draw

import "github.com/RicheyJang/PaimengBot/utils/images"

type innerItem struct {
	star int
	name string
	dir  string
	img  *images.ImageCtx
}

func simulateOnce(pool *DrawPool, last4 uint32, last5 uint32) innerItem {
	switch pool.Type {
	case PoolCharacter:
		return simulateOnceCharacter(pool, last4, last5)
	case PoolWeapon:
		return simulateOnceWeapon(pool, last4, last5)
	default:
		return simulateOnceNormal(pool, last4, last5)
	}
}

func simulateOnceNormal(pool *DrawPool, last4 uint32, last5 uint32) (item innerItem) {
	return innerItem{
		star: 5,
		name: "刻晴",
		dir:  GenshinCharacterDir,
	}
}

func simulateOnceCharacter(pool *DrawPool, last4 uint32, last5 uint32) (item innerItem) {
	return innerItem{
		star: 5,
		name: "刻晴",
		dir:  GenshinCharacterDir,
	}
}

func simulateOnceWeapon(pool *DrawPool, last4 uint32, last5 uint32) (item innerItem) {
	return innerItem{
		star: 5,
		name: "刻晴",
		dir:  GenshinCharacterDir,
	}
}
