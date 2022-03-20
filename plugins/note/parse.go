package note

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// ParseSpecTime 解析CRON时间表达式spec，填充IsOnce\Spec\RunAt
func (task *RemindTask) ParseSpecTime(spec string, isOnce bool) error {
	task.IsOnce = isOnce
	task.Spec = spec
	s, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}
	if isOnce { // 一次性
		task.RunAt = s.Next(time.Now())
		if task.RunAt.Before(time.Now()) {
			return errAlreadyPassed
		}
	}
	return nil
}

var patches = []*regexp.Regexp{
	/* 0 */ regexp.MustCompile(`^(?:今天)?(\d{1,2})[点:](\d{0,2})分?$`),
	/* 1 */ regexp.MustCompile(`^([1-9]\d{0,5})分钟后$`),
	/* 2 */ regexp.MustCompile(`^每([1-9]\d{0,5})分钟$`),
	/* 3 */ regexp.MustCompile(`^每(\d{0,4})(?:小时|钟头|个小时|个钟头)$`),

	/* 4 */ regexp.MustCompile(`^每天(\d{1,2})[点:](\d{0,2})分?$`),
	/* 5 */ regexp.MustCompile(`^(明天|后天|大后天)(\d{1,2})[点:](\d{0,2})分?$`),
	/* 6 */ regexp.MustCompile(`^([1-9]\d{0,2})天后(\d{1,2})[点:](\d{0,2})分?$`),

	/* 7 */ regexp.MustCompile(`^每个?(?:周|星期|礼拜)([1-7一二三四五六日天])(\d{1,2})[点:](\d{0,2})分?$`),
	/* 8 */ regexp.MustCompile(`^(?:周|星期|礼拜)([1-7一二三四五六日天])(\d{1,2})[点:](\d{0,2})分?$`),
	/* 9 */ regexp.MustCompile(`^每个?月([1-9]\d?)[号|日](\d{1,2})[点:](\d{0,2})分?$`),

	/* 10 */ regexp.MustCompile(`^每[个|年]([1-9]\d?)月([1-9]\d?)[号|日](\d{1,2})[点:](\d{0,2})分?$`),
	/* 11 */ regexp.MustCompile(`^([1-9]\d?)月([1-9]\d?)[号|日](\d{1,2})[点:](\d{0,2})分?$`),
}

// ParseCNTime 解析中文时间表达str，填充IsOnce\Spec\RunAt
func (task *RemindTask) ParseCNTime(str string) (err error) {
	// 匹配
	var subs []string
	var index = -1
	for i, reg := range patches {
		if reg == nil {
			continue
		}
		if subs = reg.FindStringSubmatch(str); len(subs) > reg.NumSubexp() {
			index = i
			break
		}
	}
	// 补全Task
	now := time.Now()
	switch index {
	case 0:
		task.IsOnce = true
		task.RunAt = time.Date(now.Year(), now.Month(), now.Day(), mustParseInt(subs[1]), mustParseInt(subs[2]), 0, 0, time.Local)
	case 1:
		task.IsOnce = true
		task.RunAt = now.Add(time.Duration(mustParseInt(subs[1])) * time.Minute)
	case 2:
		task.IsOnce = false
		task.Spec = fmt.Sprintf("@every %s", time.Duration(mustParseInt(subs[1]))*time.Minute)
	case 3:
		task.IsOnce = false
		hours := mustParseInt(subs[1])
		if hours <= 0 {
			hours = 1
		}
		task.Spec = fmt.Sprintf("@every %s", time.Duration(hours)*time.Hour)
	case 4:
		task.IsOnce = false
		task.Spec = fmt.Sprintf("%d %d * * *", mustParseInt(subs[2]), mustParseInt(subs[1]))
	case 5:
		task.IsOnce = true
		after := now.AddDate(0, 0, 1)
		if subs[1] == "后天" {
			after = now.AddDate(0, 0, 2)
		} else if subs[1] == "大后天" {
			after = now.AddDate(0, 0, 3)
		}
		task.RunAt = time.Date(after.Year(), after.Month(), after.Day(), mustParseInt(subs[2]), mustParseInt(subs[3]), 0, 0, time.Local)
	case 6:
		task.IsOnce = true
		after := now.AddDate(0, 0, mustParseInt(subs[1]))
		task.RunAt = time.Date(after.Year(), after.Month(), after.Day(), mustParseInt(subs[2]), mustParseInt(subs[3]), 0, 0, time.Local)
	case 7:
		task.IsOnce = false
		task.Spec = fmt.Sprintf("%d %d * * %d", mustParseInt(subs[3]), mustParseInt(subs[2]), parseWeekDay(subs[1]))
	case 8:
		task.IsOnce = true
		addDay := (parseWeekDay(subs[1]) - int(now.Weekday()) + 7) % 7
		after := now.AddDate(0, 0, addDay)
		task.RunAt = time.Date(after.Year(), after.Month(), after.Day(), mustParseInt(subs[2]), mustParseInt(subs[3]), 0, 0, time.Local)
	case 9:
		task.IsOnce = false
		task.Spec = fmt.Sprintf("%d %d %d * *", mustParseInt(subs[3]), mustParseInt(subs[2]), mustParseInt(subs[1]))
	case 10:
		task.IsOnce = false
		task.Spec = fmt.Sprintf("%d %d %d %d *", mustParseInt(subs[4]), mustParseInt(subs[3]), mustParseInt(subs[2]), mustParseInt(subs[1]))
	case 11:
		task.IsOnce = true
		task.RunAt = time.Date(now.Year(), time.Month(mustParseInt(subs[1])), mustParseInt(subs[2]), mustParseInt(subs[3]), mustParseInt(subs[4]), 0, 0, time.Local)
		if task.RunAt.Before(now) { // 若为过去日期，则加一年
			task.RunAt = task.RunAt.AddDate(1, 0, 0)
		}
	default:
		return errNoMatched
	}
	if task.IsOnce && task.RunAt.Before(time.Now()) {
		return errAlreadyPassed
	}
	return
}

func mustParseInt(s string) int {
	if len(s) == 0 {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

func parseWeekDay(s string) int {
	switch s {
	case "一":
		return 1
	case "二":
		return 2
	case "三":
		return 3
	case "四":
		return 4
	case "五":
		return 5
	case "六":
		return 6
	case "七", "日", "天":
		return 0
	}
	i := mustParseInt(s)
	if i < 0 || i >= 7 {
		return 0
	}
	return i
}
