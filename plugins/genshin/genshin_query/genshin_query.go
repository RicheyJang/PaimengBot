package genshin_query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func Query(uid string, cookie string, showLeft bool) (message.Message, *mihoyo.GenshinDailyNote, error) {
	// 查询角色信息
	role, err := mihoyo.GetUserGameRoleByUid(cookie, uid)
	if err != nil {
		return message.Message{message.Text("获取角色信息失败")}, nil, err
	}
	// 查询当前便笺
	dailyNote, err := mihoyo.GetGenshinDailyNote(cookie, uid, role.Region)
	if err != nil {
		return message.Message{message.Text("获取当前便笺失败")}, nil, err
	}
	// 构造消息
	now := time.Now()
	msg := fmt.Sprintf(
		"角色:%s(UID %v)\n[树脂:%d/%d]\n%s[洞天宝钱:%d/%d]\n%s[派遣%d/%d]",
		role.NickName, uid,
		dailyNote.CurrentResin, dailyNote.MaxResin,
		displayTime(now, dailyNote.ResinRecoveryTime, showLeft),
		dailyNote.CurrentHomeCoin, dailyNote.MaxHomeCoin,
		displayTime(now, dailyNote.HomeCoinRecoveryTime, showLeft),
		dailyNote.CurrentExpeditionNum, dailyNote.MaxExpeditionNum)
	return message.Message{message.Text(msg)}, dailyNote, err
}

// 显示时间
func displayTime(now time.Time, duration string, showLeft bool) string {
	add, err := strconv.ParseInt(duration, 10, 64)
	if err != nil || add <= 0 {
		return ""
	}
	if !showLeft { // 恢复完成时间点
		return fmt.Sprintf("(将在%v回满)\n", addTime(now, add))
	} else { // 恢复完成剩余时间
		return fmt.Sprintf("(离回满剩余: %v)\n", countDownDisplay(add))
	}
}

func addTime(now time.Time, seconds int64) string {
	after := now.Add(time.Second * time.Duration(seconds))
	return fmt.Sprintf("%d月%d日 %02d时%02d分", after.Month(), after.Day(), after.Hour(), after.Minute())
}

func countDownDisplay(add int64) string {
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
