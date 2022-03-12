package sc

import (
	"math"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"

	"github.com/RicheyJang/PaimengBot/basic/dao"
)

// RealCoin 获取真实金额，即基础货币 * 倍率
func RealCoin(base float64) float64 {
	return base * proxy.GetConfigFloat64("coin.rate")
}

// Unit 获取货币单位
func Unit() string {
	return proxy.GetConfigString("coin.unit")
}

// BaseCoinOf 获取指定用户的基础货币数
func BaseCoinOf(id int64) float64 {
	var user dao.UserOwn
	proxy.GetDB().First(&user, id)
	return user.Wealth
}

// FavorOf 获取指定用户的好感度
func FavorOf(id int64) float64 {
	var user dao.UserOwn
	proxy.GetDB().First(&user, id)
	return user.Favor
}

// LevelAt 指定好感度对应的好感度等级和升级所需的好感度
func LevelAt(favor float64) (level int, up float64) {
	level = int(math.Ceil(favor / 25.0))
	up = math.Max(25.0*float64(level)-favor, 0)
	if favor == 0 && level == 0 { // 零级零好感度时单独指定
		up = 0.01
	}
	return
}

// AddFavor 将指定用户的好感度加add，为负数时则减
func AddFavor(id int64, add float64) (left float64, ok bool) {
	if add == 0 {
		return FavorOf(id), true
	}
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		var user dao.UserOwn
		err := tx.First(&user, id).Error
		if err != nil {
			return err
		}
		// 检查是否会变为负数
		left = user.Favor
		if left+add < 0 {
			return nil
		}
		// 更新
		left += add
		err = tx.Model(&dao.UserOwn{ID: id}).Update("favor", left).Error
		if err != nil {
			left = user.Favor
			return err
		}
		ok = true
		return nil
	})
	if err != nil {
		log.Errorf("AddFavor id=%v,add=%v err: %v", id, add, err)
	}
	return
}

// AddBaseCoin 将指定用户的！基础货币！加add，为负数时则减
func AddBaseCoin(id int64, add float64) (left float64, ok bool) {
	if add == 0 {
		return BaseCoinOf(id), true
	}
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		var user dao.UserOwn
		err := tx.First(&user, id).Error
		if err != nil {
			return err
		}
		// 检查是否会变为负数
		left = user.Wealth
		if left+add < 0 {
			return nil
		}
		// 更新
		left += add
		err = tx.Model(&dao.UserOwn{ID: id}).Update("wealth", left).Error
		if err != nil {
			left = user.Favor
			return err
		}
		ok = true
		return nil
	})
	if err != nil {
		log.Errorf("AddBaseCoin id=%v,add=%v err: %v", id, add, err)
	}
	return
}
