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
	regexp.MustCompile(`^(?:今天)?([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^([1-9]\d{0,5})分钟后$`),
	regexp.MustCompile(`^每([1-9]\d{0,5})分钟$`),

	regexp.MustCompile(`^每天([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^(明天|后天|大后天)([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^([1-9]\d{0,2})天后([1-9]\d?)[点:](\d{0,2})分?$`),

	regexp.MustCompile(`^每个?(?:周|星期|礼拜)[1-7一二三四五六日天]([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^(?:周|星期|礼拜)[1-7一二三四五六日天]([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^每个月([1-9]\d?)[号|日]([1-9]\d?)[点:](\d{0,2})分?$`),

	regexp.MustCompile(`^每[个|年]([1-9]\d?)月([1-9]\d?)[号|日]([1-9]\d?)[点:](\d{0,2})分?$`),
	regexp.MustCompile(`^([1-9]\d?)月([1-9]\d?)[号|日]([1-9]\d?)[点:](\d{0,2})分?$`),
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
	// TODO 补全Task
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
	default:
		return errNoMatched
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
