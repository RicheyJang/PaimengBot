package sc

import (
	"math"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

const ReturnCostTag = "return_cost"

// SetNeedReturnCost 标记当前操作需要返还扣除的货币
func SetNeedReturnCost(ctx *zero.Ctx) {
	ctx.State[ReturnCostTag] = true
}

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
	proxy.GetDB().Find(&user, id)
	return user.Wealth
}

// FavorOf 获取指定用户的好感度
func FavorOf(id int64) float64 {
	var user dao.UserOwn
	proxy.GetDB().Find(&user, id)
	return user.Favor
}

// LevelAt 指定好感度对应的好感度等级和升级所需的好感度
func LevelAt(favor float64) (level int, up float64) {
	level = int(math.Ceil(favor / 25.0))
	up = math.Max(SumFavorAt(level)*float64(level)-favor, 0)
	return
}

// SumFavorAt 在level级别升级所需的总好感度
func SumFavorAt(level int) float64 {
	if level == 0 { // 零级零好感度时单独指定
		return 0.01
	}
	return 25.0
}

// AddFavor 将指定用户的好感度加add，为负数时则减
func AddFavor(id int64, add float64) (left float64, ok bool) {
	if add == 0 {
		return FavorOf(id), true
	}
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		var user dao.UserOwn
		res := tx.Find(&user, id)
		if res.Error != nil {
			return res.Error
		}
		// 检查是否会变为负数
		left = user.Favor
		if left+add < 0 {
			return nil
		}
		left += add
		// 若不存在，则创建
		if res.RowsAffected == 0 {
			if err := tx.Create(&dao.UserOwn{ID: id, Favor: left}).Error; err != nil {
				left = 0
				return err
			}
			ok = true
			return nil
		}
		// 更新
		if err := tx.Model(&dao.UserOwn{ID: id}).Update("favor", left).Error; err != nil {
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
		res := tx.Find(&user, id)
		if res.Error != nil {
			return res.Error
		}
		// 检查是否会变为负数
		left = user.Wealth
		if left+add < 0 {
			return nil
		}
		left += add
		// 若不存在，则创建
		if res.RowsAffected == 0 {
			if err := tx.Create(&dao.UserOwn{ID: id, Wealth: left}).Error; err != nil {
				left = 0
				return err
			}
			ok = true
			return nil
		}
		// 更新
		if err := tx.Model(&dao.UserOwn{ID: id}).Update("wealth", left).Error; err != nil {
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
