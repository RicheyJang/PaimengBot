package dao

import (
	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
)

type UserSetting struct {
	ID           int64   `gorm:"primaryKey;autoIncrement:false"`
	BlackPlugins string  `gorm:"size:512"` // 白名单插件
	WhitePlugins string  `gorm:"size:512"` // 黑名单插件
	Nickname     string  // 昵称
	Likeability  float64 // 好感度
	Flag         string  // 非空时代表该用户尚未成为好友，是他的好友请求flag
}

type GroupSetting struct {
	ID           int64  `gorm:"primaryKey;autoIncrement:false"`
	BlackPlugins string `gorm:"size:512"`
	WhitePlugins string `gorm:"size:512"`
	Flag         string // 非空时代表该群尚未加入，是邀请入群请求flag
	CouldAdd     bool   `gorm:"default:false"` // 能否入此群标志位
	Welcome      string
}

type UserPriority struct {
	ID       int64 `gorm:"primaryKey;autoIncrement:false"`
	GroupID  int64 `gorm:"primaryKey;autoIncrement:false"`
	Priority int
}

func init() {
	err := manager.GetDB().AutoMigrate(&UserSetting{}, &GroupSetting{}, &UserPriority{})
	if err != nil {
		log.Fatalf("初始化基本数据库失败 err: %v", err)
	}
}
