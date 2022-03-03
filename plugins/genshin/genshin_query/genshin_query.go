package genshin_query

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_public"
	"strconv"
	"time"
)

func Query(uid string, cookie string) (string, *genshin_public.GenshinDailyNote, error) {
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
		"用户:%s\n[用户树脂:%d/%d]\n[树脂回满时间:%s]\n[用户洞天宝钱:%d/%d]\n[宝钱回满时间:%s]\n[用户派遣%d/%d]",
		uid,
		dailyNote.CurrentResin,
		dailyNote.MaxResin,
		addTime(now, dailyNote.ResinRecoveryTime),
		dailyNote.CurrentHomeCoin, dailyNote.MaxHomeCoin,
		addTime(now, dailyNote.HomeCoinRecoveryTime),
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
