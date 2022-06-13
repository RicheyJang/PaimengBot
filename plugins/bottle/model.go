package bottle

import (
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
)

type DriftingBottleModel struct {
	ID        int
	FromID    int64 `gorm:"column:from_id"`
	Content   string
	CreatedAt time.Time
}

func init() {
	err := manager.GetDB().AutoMigrate(&DriftingBottleModel{})
	if err != nil {
		log.Errorf("[SQL] Bottle DriftingBottleModel初始化失败，err: %v", err)
	}
}
