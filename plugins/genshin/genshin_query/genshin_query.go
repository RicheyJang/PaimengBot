package genshin_query

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_public"
	"strconv"
	"time"
)

func Query(uid string, cookie string, displayType int) (string, *genshin_public.GenshinDailyNote, error) {
	role, err := genshin_public.GetUserGameRoleByUid(cookie, uid)
	if err != nil {
		msg := fmt.Sprintf("获取角色信息失败,error:%s", err.Error())
		return msg, nil, err
	}
	dailyNote, err := genshin_public.GetGenshinDailyNote(cookie, uid, role.Region)
	if err != nil {
		msg := fmt.Sprintf("获取角色信息失败,error:%s", err.Error())
		return msg, nil, err
	}
	now := time.Now()
	msg := fmt.Sprintf(
		"用户:%s\n[用户树脂:%d/%d]\n[%s]\n[用户洞天宝钱:%d/%d]\n[%s]\n[用户派遣%d/%d]",
		uid,
		dailyNote.CurrentResin,
		dailyNote.MaxResin,
		displayTime(now, dailyNote.ResinRecoveryTime, displayType),
		dailyNote.CurrentHomeCoin, dailyNote.MaxHomeCoin,
		displayTime(now, dailyNote.HomeCoinRecoveryTime, displayType),
		dailyNote.CurrentExpeditionNum, dailyNote.MaxExpeditionNum)
	return msg, dailyNote, err
}
func addTime(now time.Time, duration string) string {
	add, err := strconv.Atoi(duration)
	if err != nil {
		after := now.Add(8 * 160 * time.Minute)
		return fmt.Sprintf("%04d-%02d-%02d %02dH-%02dM", after.Year(), after.Month(), after.Day(), after.Hour(), after.Minute())
	}
	after := now.Add(time.Duration(time.Second * time.Duration(int(add))))
	return fmt.Sprintf("%04d-%02d-%02d %02d小时%02d分", after.Year(), after.Month(), after.Day(), after.Hour(), after.Minute())
}

func CountDownDisplay(duration string) string {
	add, err := strconv.Atoi(duration)
	if err != nil {
		return "无法解析时间"
	}
	// 判断是否包含天
	if add > 86400 {
		return fmt.Sprintf("%d天%d小时%d分", add/86400, (add%86400)/3600, (add%3600)/60)
	}
	// 判断是否包含小时
	if add > 3600 {
		return fmt.Sprintf("%d小时%d分", add/3600, (add%3600)/60)
	}
	// 判断是否包含分钟
	if add > 60 {
		return fmt.Sprintf("%d分", add/60)
	}
	return fmt.Sprintf("%d秒", add)
}

// 显示时间
func displayTime(now time.Time, duration string, displayType int) string {

	if displayType == 1 {
		return "rct:" + addTime(now, duration)
	} else if displayType == 2 {
		// use CountDownDisplay
		return "离回满剩余:" + CountDownDisplay(duration)
	} else {
		// default method is 1
		return "rct:" + addTime(now, duration)
	}
	// undefined
	return "无法解析时间"
}
