package genshin_query

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin"
	"strconv"
	"time"
)

func Query(uid string, cookie string) (string, *genshin.GenshinDailyNote, error) {
	role, err := genshin.GetUserGameRoleByUid(cookie, uid)
	if err != nil {
		msg := fmt.Sprintf("获取角色信息失败,error:%s", err.Error())
		return msg, nil, err
	}
	dailyNote, err := genshin.GetGenshinDailyNote(cookie, uid, role.Region)
	if err != nil {
		msg := fmt.Sprintf("获取角色信息失败,error:%s", err.Error())
		return msg, nil, err
	}
	now := time.Now()
	msg := fmt.Sprintf(
		"用户:%s\n[用户树脂:%d/%d]\n[树脂回满时间:%s]\n[用户洞天宝钱:%d/%d]\n[用户派遣%d/%d]",
		uid,
		dailyNote.CurrentResin,
		dailyNote.MaxResin,
		addTime(now, dailyNote.HomeCoinRecoveryTime), dailyNote.CurrentHomeCoin, dailyNote.MaxHomeCoin, dailyNote.CurrentExpeditionNum, dailyNote.MaxExpeditionNum)
	return msg, dailyNote, err
}
func addTime(now time.Time, duration string) string {
	add, err := strconv.Atoi(duration)
	if err != nil {
		after := now.Add(8 * 160 * time.Minute)
		return after.String()
	}
	after := now.Add(time.Duration(time.Second * time.Duration(add/10)))
	return after.String()
}
