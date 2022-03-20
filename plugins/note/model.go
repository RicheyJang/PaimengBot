package note

import (
	"fmt"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/push"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

func init() {
	err := manager.GetDB().AutoMigrate(&RemindTask{})
	if err != nil {
		log.Errorf("[SQL] Note RemindTask初始化失败，err: %v", err)
	}
}

type RemindTask struct {
	ID      int64
	UserID  int64
	GroupID int64
	Content string
	// 根据用户设定生成，一次性写入
	IsOnce bool      // 是否为一次性任务
	Spec   string    // 重复性任务专用：代表CRON表达式
	RunAt  time.Time // 一次性任务专用：执行时间点
	// 程序维护
	CronID    int // 唯一的每次重启bot都需重写的字段
	CreatedAt time.Time
}

// Job 生成 到达定时点时所要执行的 任务函数
func (task RemindTask) Job() func() {
	return func() {
		target := push.Target{
			Msg: message.ParseMessageFromString(task.Content),
		}
		if task.GroupID != 0 { // 群推送
			target.Groups = append(target.Groups, task.GroupID)
		} else { // 个人推送
			target.Friends = append(target.Friends, task.UserID)
		}
		target.Send()
	}
}

// 添加定时任务，并修改数据库
func (task *RemindTask) addJob(scheduler cron.Schedule) (err error) {
	if task == nil || scheduler == nil {
		return fmt.Errorf("wrong param")
	}
	fn := task.Job()
	// 添加定时任务
	var id cron.EntryID
	if task.IsOnce {
		id, err = proxy.AddSchedule(scheduler, func() {
			proxy.DeleteSchedule(id) // 执行一次后删除
			go cleanIllegalTasks()
			fn()
		})
	} else {
		id, err = proxy.AddSchedule(scheduler, fn)
	}
	task.CronID = int(id)
	// 将task添加进数据库或更新cronID
	err = proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cron_id"}), // Upsert
	}).Create(task).Error
	return
}

// 清除超时的数据库任务
func cleanIllegalTasks() {
	var tasks []RemindTask
	proxy.GetDB().Find(&tasks)
	for _, task := range tasks {
		if _, err := genSchedule(task); err != nil {
			log.Infof("删除无效的提醒任务，ID=%v, 原因=%v", task.ID, err)
			if proxy.GetDB().Delete(&task).Error != nil {
				log.Errorf("[SQL] 无法删除超时的提醒任务，ID=%v", task.ID)
			}
		}
	}
}
