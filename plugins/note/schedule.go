package note

import (
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// 依据task生成Schedule
func genSchedule(task RemindTask) (scheduler cron.Schedule, err error) {
	startAt := time.Now()
	if !task.CreatedAt.IsZero() {
		startAt = task.CreatedAt
	}
	if task.IsOnce { // 一次性
		if task.RunAt.Before(time.Now()) {
			return nil, errAlreadyPassed
		}
		return StickTimeSchedule{At: task.RunAt}, nil
	} else { // 重复性
		// 根据startAt处理特殊情况@every
		if strings.HasPrefix(task.Spec, "@every") {
			periodStr := strings.TrimSpace(task.Spec[len("@every"):])
			var period time.Duration
			period, err = time.ParseDuration(periodStr)
			if err != nil {
				return nil, err
			}
			// 使用自定义的ConstantEverySchedule
			return ConstantEverySchedule{
				StartAt: startAt,
				Delay:   period,
			}, nil
		}
		// 其它Cron
		return cron.ParseStandard(task.Spec)
	}
}

// ConstantEverySchedule 重复性间隔时间执行，StartAt代表开始时间点
type ConstantEverySchedule struct {
	StartAt time.Time
	Delay   time.Duration
}

func (schedule ConstantEverySchedule) Next(t time.Time) time.Time {
	if schedule.StartAt.Add(schedule.Delay).After(t) {
		t = schedule.StartAt
	}
	return t.Add(schedule.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
}

// StickTimeSchedule 固定单一时间点执行
type StickTimeSchedule struct {
	At time.Time
}

func (schedule StickTimeSchedule) Next(_ time.Time) time.Time {
	return schedule.At
}
