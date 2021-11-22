package dao

import (
	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
)

type UserSetting struct {
	ID           int64  `gorm:"primaryKey;autoIncrement:false"`
	BlackPlugins string `gorm:"size:512"`
	WhitePlugins string `gorm:"size:512"`
	NickName     string
	Likeability  float64
}

type GroupSetting struct {
	ID           int64  `gorm:"primaryKey;autoIncrement:false"`
	BlackPlugins string `gorm:"size:512"`
	WhitePlugins string `gorm:"size:512"`
}

type UserPriority struct {
	ID       int64 `gorm:"primaryKey;autoIncrement:false"`
	GroupID  int64 `gorm:"primaryKey;autoIncrement:false"`
	Priority int
}

func init() {
	err := manager.GetDB().AutoMigrate(&UserSetting{}, &GroupSetting{}, &UserPriority{})
	if err != nil {
		log.Errorf("init basic models err: %v", err)
	}
}
