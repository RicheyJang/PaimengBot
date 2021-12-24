package genshin_draw

import (
	"math/rand"

	"github.com/RicheyJang/PaimengBot/utils/images"
)

type innerItem struct {
	star int
	name string

	img *images.ImageCtx
}

// 概率

var percentOfNC5 = map[uint32]float64{
	0:  0.60,
	74: 6.60,
}

var percentOfNC4 = map[uint32]float64{
	0:  5.10,
	9:  56.10,
	10: 100,
	11: 100,
}

var percentOfW5 = map[uint32]float64{
	0:  0.70,
	63: 7.70,
}

var percentOfW4 = map[uint32]float64{
	0:  6,
	8:  66,
	9:  96,
	10: 100,
	11: 100,
}

func init() {
	var i uint32
	for i = 1; i <= 90; i++ {
		if i < 74 {
			percentOfNC5[i] = percentOfNC5[0]
		} else if i > 74 {
			percentOfNC5[i] = percentOfNC5[i-1] + 6.0
		}
	}
	for i = 1; i <= 8; i++ {
		percentOfNC4[i] = percentOfNC4[0]
	}
	for i = 1; i <= 80; i++ {
		if i < 63 {
			percentOfW5[i] = percentOfW5[0]
		} else if i > 63 {
			percentOfW5[i] = percentOfW5[i-1] + 7.0
		}
	}
	for i = 1; i <= 7; i++ {
		percentOfW4[i] = percentOfW4[0]
	}
}

// 模拟抽卡

func simulateOnce(pool *DrawPool, user *UserInfo) innerItem {
	switch pool.Type {
	case PoolCharacter:
		return simulateOnceCharacter(pool, user)
	case PoolWeapon:
		return simulateOnceWeapon(pool, user)
	default:
		return simulateOnceNormal(pool, user)
	}
}

func simulateOnceNormal(pool *DrawPool, user *UserInfo) (item innerItem) {
	item.star = randomStar(getFloatRate(percentOfNC4, user.Last4), getFloatRate(percentOfNC5, user.Last5))
	user.Last4 += 1
	user.Last5 += 1
	switch item.star {
	case 3:
		item.name = randomChoice(pool.Normal3)
	case 4:
		user.Last4 = 0
		item.name = randomChoice(pool.Normal4)
	case 5:
		user.Last5 = 0
		if randomZeroOne(0.5) {
			item.name = randomChoice(pool.Normal5Weapon)
		} else {
			item.name = randomChoice(pool.Normal5Character)
		}
	}
	return
}

func simulateOnceCharacter(pool *DrawPool, user *UserInfo) (item innerItem) {
	item.star = randomStar(getFloatRate(percentOfNC4, user.CLast4), getFloatRate(percentOfNC5, user.CLast5))
	user.CLast4 += 1
	user.CLast5 += 1
	switch item.star {
	case 3:
		item.name = randomChoice(pool.Normal3)
	case 4:
		user.CLast4 = 0
		if user.C4LastUp > 0 || randomZeroOne(0.5) { // UP
			item.name = randomChoice(pool.Limit4)
			user.C4LastUp = 0
		} else {
			item.name = randomChoice(pool.Normal4)
			user.C4LastUp += 1
		}
	case 5:
		user.CLast5 = 0
		if user.C5LastUp > 0 || randomZeroOne(0.5) { // UP
			item.name = randomChoice(pool.Limit5)
			user.C5LastUp = 0
		} else {
			item.name = randomChoice(pool.Normal5Character)
			user.C5LastUp += 1
		}
	}
	return
}

func simulateOnceWeapon(pool *DrawPool, user *UserInfo) (item innerItem) {
	item.star = randomStar(getFloatRate(percentOfW4, user.WLast4), getFloatRate(percentOfW5, user.WLast5))
	user.WLast4 += 1
	user.WLast5 += 1
	switch item.star {
	case 3:
		item.name = randomChoice(pool.Normal3)
	case 4:
		user.WLast4 = 0
		if user.W4LastUp > 0 || randomZeroOne(0.75) { // UP
			item.name = randomChoice(pool.Limit4)
			user.W4LastUp = 0
		} else {
			item.name = randomChoice(pool.Normal4)
			user.W4LastUp += 1
		}
	case 5:
		user.WLast5 = 0
		if user.W5LastUp > 0 || randomZeroOne(0.75) { // UP
			// TODO 武器定轨
			item.name = randomChoice(pool.Limit5)
			user.W5LastUp = 0
		} else {
			item.name = randomChoice(pool.Normal5Weapon)
			user.W5LastUp += 1
		}
	}
	return
}

// 抽卡统一事后处理
func (user *UserInfo) postProcess(pool *DrawPool, item innerItem) {
	// 暂时没啥统一的
}

func getFloatRate(mp map[uint32]float64, last uint32) float64 {
	rate, ok := mp[last+1]
	if !ok { // 必出，防止数据错误
		return 1.1
	}
	return rate / 100.0
}

// 随机函数

func randomStar(rateOf4, rateOf5 float64) int {
	if rateOf5 >= 1 {
		return 5
	}
	if rateOf4 >= 1 {
		return 4
	}
	if rateOf4+rateOf5 >= 1 {
		rateOf4, rateOf5 = rateOf4/(rateOf4+rateOf5), rateOf5/(rateOf4+rateOf5)
	}
	rateOf5 = rateOf4 + rateOf5
	r := rand.Float64()
	if r < rateOf4 {
		return 4
	} else if r < rateOf5 {
		return 5
	}
	return 3
}

func randomChoice(src []string) string {
	if len(src) == 0 {
		return ""
	}
	return src[rand.Intn(len(src))]
}

func randomZeroOne(rateTrue float64) bool {
	r := rand.Float64()
	if r < rateTrue {
		return true
	}
	return false
}
