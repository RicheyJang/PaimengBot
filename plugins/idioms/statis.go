package idioms

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/images"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

func rankHandler(ctx *zero.Ctx) {
	// 查询
	var users []GuessIdiomsModel
	if err := proxy.GetDB().Model(&GuessIdiomsModel{}).Select("id, SUM(success) as success").
		Group("id").Order("success desc").Limit(10).Find(&users).Error; err != nil {
		log.Errorf("查询失败：%v", err)
		ctx.Send("失败了...")
		return
	}
	if len(users) == 0 {
		ctx.Send("暂时没人猜对过成语")
		return
	}
	// 画图
	var values []images.UserValue
	for _, user := range users {
		values = append(values, images.UserValue{
			ID:    user.ID,
			Value: float64(user.Success),
		})
	}
	msg, _ := images.GenQQRankMsgWithValue("猜成语总排行榜", values, "个")
	ctx.Send(msg)
}

// 记录某人猜对了成语
func addSuccess(group, id int64) {
	fErr := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		guess := GuessIdiomsModel{
			ID:      id,
			GroupID: group,
		}
		res := tx.Find(&guess)
		if err := res.Error; err != nil {
			return err
		}
		if res.RowsAffected == 0 { // 创建
			guess.Success = 1
			if err := tx.Create(&guess).Error; err != nil {
				return err
			}
			return nil
		}
		if err := tx.Model(&guess).Update("success", guess.Success+1).Error; err != nil { // 更新
			return err
		}
		return nil
	})
	if fErr != nil {
		log.Warnf("猜成语记录失败：%v", fErr)
	}
}

type GuessIdiomsModel struct {
	ID      int64 `gorm:"primaryKey;autoIncrement:false"`
	GroupID int64 `gorm:"primaryKey;autoIncrement:false"`
	Success int
}

func init() {
	err := manager.GetDB().AutoMigrate(&GuessIdiomsModel{})
	if err != nil {
		log.Errorf("[SQL] Idioms GuessIdiomsModel初始化失败，err: %v", err)
	}
}
