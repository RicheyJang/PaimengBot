package pixiv

import (
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
)

// OmegaPixivIllusts Omega图库结构，请从https://github.com/Ailitonia/omega-miya/raw/master/archive_data/db_pixiv.7z手动导入数据库
// 特别鸣谢Ailitonia/omega-miya项目
type OmegaPixivIllusts struct {
	ID        int       `gorm:"column:id;primaryKey"`
	PID       int64     `gorm:"column:pid;uniqueIndex:ix_omega_pixiv_illusts_pid"`
	UID       int64     `gorm:"column:uid"`
	Title     string    `gorm:"column:title"`
	Uname     string    `gorm:"column:uname"`
	NsfwTag   int       `gorm:"column:nsfw_tag"` //nsfw标签, 0=safe, 1=setu, 2=r18
	Width     int       `gorm:"column:width"`
	Height    int       `gorm:"column:height"`
	Tags      string    `gorm:"column:tags;size:1024"`
	URL       string    `gorm:"column:url;size:1024"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (p OmegaPixivIllusts) TableName() string {
	return "omega_pixiv_illusts"
}

func init() {
	err := manager.GetDB().AutoMigrate(&OmegaPixivIllusts{})
	if err != nil {
		log.Errorf("[SQL] OmegaPixivIllusts 初始化失败, err: %v", err)
	}
}
