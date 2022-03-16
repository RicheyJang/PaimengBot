package dao

import (
	"time"

	"github.com/RicheyJang/PaimengBot/manager"

	log "github.com/sirupsen/logrus"
)

type UserSetting struct {
	ID           int64   `gorm:"primaryKey;autoIncrement:false"`
	BlackPlugins string  `gorm:"size:512"` // 白名单插件
	WhitePlugins string  `gorm:"size:512"` // 黑名单插件
	Nickname     string  // 昵称
	Likeability  float64 // 好感度（无用，抱歉...）
	Flag         string  // 非空时代表该用户尚未成为好友，是他的好友请求flag
	IsPullBlack  bool    // 是否被拉黑：拒绝好友请求、拒绝加群请求
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

type UserOwn struct {
	ID       int64     `gorm:"primaryKey;autoIncrement:false"`
	Favor    float64   // 好感度
	LastSign time.Time // 上次签到时间
	SignDays int       // 连续签到天数
	Wealth   float64   // 拥有的基础货币数量
	Items    string    // 拥有的物品列表，商店相关
}

func init() {
	err := manager.GetDB().AutoMigrate(&UserSetting{}, &GroupSetting{}, &UserPriority{}, &UserOwn{})
	if err != nil {
		log.Fatalf("初始化基本数据库失败 err: %v", err)
	}
}
