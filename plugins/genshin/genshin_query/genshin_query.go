package genshin_query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/wdvxdr1123/ZeroBot/message"
)

func Query(uid string, cookie string, showLeft bool) (message.Message, *mihoyo.GenshinDailyNote, error) {
	// 查询角色信息
	role, err := mihoyo.GetUserGameRoleByUid(cookie, uid)
	if err != nil {
		return message.Message{message.Text("获取角色信息失败")}, nil, err
	}
	// 查询当前便签
	dailyNote, err := mihoyo.GetGenshinDailyNote(cookie, uid, role.Region)
	if err != nil {
		return message.Message{message.Text("获取当前便签失败")}, nil, err
	}
	// 构造消息
	msg, err := genNotePicMessage(role, dailyNote, showLeft)
	if err != nil { // 图片生成失败，使用文字
		now := time.Now()
		str := fmt.Sprintf(
			"角色:%s(UID %v)\n[树脂:%d/%d]\n(%s)\n[洞天宝钱:%d/%d]\n(%s)\n[派遣%d/%d]",
			role.NickName, uid,
			dailyNote.CurrentResin, dailyNote.MaxResin,
			displayTime(now, dailyNote.ResinRecoveryTime, "回满", showLeft),
			dailyNote.CurrentHomeCoin, dailyNote.MaxHomeCoin,
			displayTime(now, dailyNote.HomeCoinRecoveryTime, "回满", showLeft),
			dailyNote.CurrentExpeditionNum, dailyNote.MaxExpeditionNum)
		return message.Message{message.Text(str)}, dailyNote, err
	}
	return message.Message{msg}, dailyNote, nil
}

// 生成图片便签消息
func genNotePicMessage(role *mihoyo.GameRole, note *mihoyo.GenshinDailyNote, showLeft bool) (message.MessageSegment, error) {
	now := time.Now()
	maxExpedition := "0"
	for _, r := range note.Expeditions {
		if r.RemainedTime > maxExpedition {
			maxExpedition = r.RemainedTime
		}
	}
	img := images.NewImageCtxWithBGColor(680, 600, "#f5eee6")
	// 角色
	err := img.PasteStringDefault(role.NickName, 28, 1, 50, 15, 680)
	if err != nil {
		return message.MessageSegment{}, err
	}
	// 设置参数
	if err = img.UseDefaultFont(24); err != nil {
		return message.MessageSegment{}, err
	}
	height := 60.0
	// 计算文字
	lineUp := [5]string{"原粹树脂", "洞天财瓮 - 洞天宝钱", "每日委托任务", "值得铭记的强敌", "探索派遣"}
	lineDown := [5]string{displayTime(now, note.ResinRecoveryTime, "回满", showLeft),
		displayTime(now, note.HomeCoinRecoveryTime, "回满", showLeft),
		"还有委托没打！", "本周剩余消耗减半次数",
		displayTime(now, maxExpedition, "全部完成", showLeft)}
	if note.FinishedTaskNum == note.TotalTaskNum {
		lineDown[2] = "委托已全部完成"
		if !note.GetTaskExReward {
			lineDown[2] += "，但还没有领取额外奖励！"
		}
	}
	if note.CurrentExpeditionNum == 0 {
		lineDown[4] = "还没有派遣"
	}
	countLeft := [5]int{note.CurrentResin, note.CurrentHomeCoin, note.FinishedTaskNum, note.RemainResinDiscountNum, note.CurrentExpeditionNum}
	countRight := [5]int{note.MaxResin, note.MaxHomeCoin, note.TotalTaskNum, note.ResinDiscountNumLimit, note.MaxExpeditionNum}
	// 画图
	for i := range lineUp {
		img.PasteRectangle(30, height, 620, 90, "#ebe3d8")
		img.PasteRectangle(32, height+2, 620-2-156, 90-4, "#f5f2eb")
		img.SetColorAuto("#62615d")
		img.DrawStringAnchored(lineUp[i], 50, height+45, 0, -0.2)
		img.SetColorAuto("#c3b4a4")
		img.DrawStringAnchored(lineDown[i], 50, height+45, 0, 1)
		img.SetColorAuto("#805e42")
		img.DrawStringAnchored(fmt.Sprintf("%d/%d", countLeft[i], countRight[i]),
			30+620-78, height+45, 0.5, 0.5)
		height += 90 + 10
	}
	img.SetHexColor("#a6a19c")
	img.DrawStringAnchored("UID "+role.Uid, 630, height, 1, 1)
	return img.GenMessageAuto()
}

// 显示时间
func displayTime(now time.Time, duration string, finishTip string, showLeft bool) string {
	add, err := strconv.ParseInt(duration, 10, 64)
	if err != nil || add <= 0 {
		return "已" + finishTip
	}
	if !showLeft { // 恢复完成时间点
		return fmt.Sprintf("将在%v%s", addTime(now, add), finishTip)
	} else { // 恢复完成剩余时间
		return fmt.Sprintf("距%s还有: %v", finishTip, countDownDisplay(add))
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
