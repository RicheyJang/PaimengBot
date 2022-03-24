package note

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "定时提醒",
	Usage: `用法：
	[时间]提醒[我 / 本群 / 群+ID] [提醒内容]：为自己或者指定群设置定时提醒，设置群提醒时需要你是该群的管理员
	定时提醒：私聊中，查看你所设置的所有定时提醒；群聊中，查看该群的所有定时提醒（需为管理员）
	取消提醒 [事件ID]：事件ID为"定时提醒"中所展示的
示例：
	每10分钟提醒我快去学习`, // TODO 示例
	SuperUsage: `config-plugin配置项：
	note.max: 单个用户的定时提醒配额上限`,
}
var proxy *manager.PluginProxy
var errNoMatched = fmt.Errorf("no regex matched")
var errAlreadyPassed = fmt.Errorf("already passed")

const MinSetGroupNoteLevel = 5

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnFullMatch([]string{"定时提醒", "定时提醒列表", "定时提醒清单"}).SetBlock(true).ThirdPriority().Handle(listNoteHandler)
	proxy.OnCommands([]string{"取消提醒", "取消定时提醒"}).SetBlock(true).ThirdPriority().Handle(cancelNoteHandler)
	proxy.OnRegex(`^(.+)提醒(我|本群|群\d+)(.*)`, zero.OnlyToMe).SetBlock(true).SetPriority(3).Handle(noteHandler)
	proxy.AddConfig("max", 6)
	initAllTasks()
}

func noteHandler(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) <= 3 {
		ctx.Send("参数不足哦，可以看看帮助")
		return
	}
	var task RemindTask
	var err error
	// 解析ID、鉴权
	task.UserID = ctx.Event.UserID
	groupID := ctx.Event.GroupID
	if subs[2] == "我" && groupID != 0 {
		ctx.Send("请在私聊中设置")
		return
	} else if strings.HasPrefix(subs[2], "群") {
		groupID, err = strconv.ParseInt(subs[2][len("群"):], 10, 64)
		if err != nil {
			ctx.Send("群ID格式不对哦")
			return
		}
	} else if subs[2] == "本群" && groupID == 0 {
		ctx.Send("请在群聊中设置")
		return
	}
	if MinSetGroupNoteLevel > 0 && groupID != 0 {
		level := auth.GetGroupUserPriority(groupID, ctx.Event.UserID)
		if utils.IsGroupAnonymous(ctx) { // 匿名消息，权限设为最低
			level = math.MaxInt
		}
		if level > MinSetGroupNoteLevel {
			ctx.Send(fmt.Sprintf("你的权限不足喔，需要权限%v", MinSetGroupNoteLevel))
			return
		}
	}
	task.GroupID = groupID
	// 检查配额
	var count int64
	proxy.GetDB().Model(&RemindTask{}).Where(&RemindTask{UserID: ctx.Event.UserID}).Count(&count)
	if count >= proxy.GetConfigInt64("max") && !utils.IsSuperUser(ctx.Event.UserID) {
		log.Infof("用户%d定时提醒任务数量(%d)达到配额上限，无法继续配置", ctx.Event.UserID, count)
		ctx.Send(fmt.Sprintf("达到配额上限(%d)，无法配置更多的定时提醒了，可以取消一些或者联系一下管理员试试", proxy.GetConfigInt64("max")))
		return
	}
	// 解析时间
	if strings.Count(subs[1], " ") == 4 || strings.HasPrefix(subs[1], "@") {
		err = task.ParseSpecTime(subs[1], false)
	} else {
		err = task.ParseCNTime(subs[1])
	}
	if err != nil {
		log.Errorf("Parse Time(%v) err : %v", subs[1], err)
		ctx.Send("时间格式不对哦")
		return
	}
	// 尝试生成定时器
	s, err := task.genSchedule()
	if err != nil {
		log.Errorf("genSchedule(%v) err : %v", subs[1], err)
		if err == errAlreadyPassed {
			ctx.Send("不能用已经过去的时间")
		} else {
			ctx.Send("时间格式不对哦")
		}
		return
	}
	// 内容
	content := ctx.Event.Message.String() // 使用消息的全部剩余部分
	index := strings.Index(content, "提醒"+subs[2])
	task.Content = content[index+len("提醒"+subs[2]):]
	if len(task.Content) == 0 {
		ctx.Send("提醒内容是什么？直接说即可")
		e := utils.WaitNextMessage(ctx)
		if e != nil {
			task.Content = e.Message.String()
		} else {
			return
		}
	}
	// 添加任务
	if err = task.addJob(s); err != nil {
		log.Errorf("addJob failed: %v", err)
		ctx.Send("失败了...")
		return
	}
	// 剩下的
	ctx.Send("成功添加新提醒\n" + task.String())
}

func cancelNoteHandler(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		ctx.Send(`事件ID格式不对，请以"定时提醒"中的事件ID为准`)
		return
	}
	// 获取任务
	var task RemindTask
	if proxy.GetDB().First(&task, id).RowsAffected == 0 {
		ctx.Send("不存在该事件ID")
		return
	}
	// 鉴权
	if MinSetGroupNoteLevel > 0 && task.GroupID != 0 {
		level := auth.GetGroupUserPriority(task.GroupID, ctx.Event.UserID)
		if utils.IsGroupAnonymous(ctx) { // 匿名消息，权限设为最低
			level = math.MaxInt
		}
		if level > MinSetGroupNoteLevel {
			ctx.Send(fmt.Sprintf("你的权限不足喔，需要权限%v", MinSetGroupNoteLevel))
			return
		}
	}
	if task.GroupID == 0 && task.UserID != ctx.Event.UserID && !utils.IsSuperUser(ctx.Event.UserID) {
		ctx.Send("只能取消你自己设置的定时提醒哦")
		return
	}
	// 删除定时任务
	proxy.DeleteSchedule(cron.EntryID(task.CronID))
	// 数据库删除
	if err = proxy.GetDB().Delete(&RemindTask{}, id).Error; err != nil {
		log.Errorf("[SQL] Delete err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("好哒")
}

func listNoteHandler(ctx *zero.Ctx) {
	var err error
	var tasks []RemindTask
	// 获取所有任务
	if ctx.Event.GroupID != 0 && auth.CheckPriority(ctx, MinSetGroupNoteLevel, false) {
		// 群管所在的群
		err = proxy.GetDB().Where(&RemindTask{GroupID: ctx.Event.GroupID}).Find(&tasks).Error
	} else {
		// 普通用户自行设置
		err = proxy.GetDB().Where(&RemindTask{
			UserID:  ctx.Event.UserID,
			GroupID: ctx.Event.GroupID,
		}).Find(&tasks).Error
	}
	if err != nil {
		log.Errorf("[SQL] get some RemindTasks err: %v", err)
		ctx.Send("失败了...")
		return
	}
	if len(tasks) == 0 {
		ctx.Send("你暂时没有设置过定时提醒")
		return
	}
	// 生成消息
	for _, task := range tasks {
		ctx.Send(task.String())
		time.Sleep(time.Millisecond * 100)
	}
}

func getTaskTimeList(task RemindTask) string {
	s, err := task.genSchedule()
	if err == errAlreadyPassed {
		return "已经过预定时间"
	} else if err != nil || s == nil {
		return "未知"
	}
	next := s.Next(time.Now())
	if entry := proxy.GetScheduleEntry(cron.EntryID(task.CronID)); entry.ID != 0 {
		next = entry.Next
	}
	if next.IsZero() {
		return "无效时间"
	}
	str := genBriefTime(next)
	if task.IsOnce {
		return str
	}
	next = s.Next(next.Add(time.Second))
	if !next.IsZero() {
		str += "," + genBriefTime(next) + "..."
	}
	return str
}

func genBriefTime(t time.Time) string {
	str := t.Format("15:04")
	now := time.Now()
	if now.Year() != t.Year() {
		str = t.Format("2006年01月02日") + str
	} else if now.Month() != t.Month() || now.Day() != t.Day() {
		str = t.Format("01月02日") + str
	}
	return str
}

func genBriefMessage(msg message.Message) string {
	// 含有文字
	str := msg.ExtractPlainText()
	if len(str) > 0 {
		runeOfStr := []rune(str)
		if len(runeOfStr) > 10 {
			str = string(runeOfStr[10:]) + "..."
		}
		return str
	}
	// 不含文字，挑选第一个不是@的消息类型
	var selectSeg message.MessageSegment
	for _, seg := range msg {
		selectSeg = seg
		if seg.Type != "at" {
			break
		}
	}
	// 文字
	switch selectSeg.Type {
	case "face":
		return selectSeg.String() + "..."
	case "image":
		return "图片"
	case "record":
		return "一段语言"
	case "video":
		return "一段视频"
	case "at":
		return "@" + selectSeg.Data["qq"]
	case "share":
		return "分享"
	default:
		return selectSeg.Type + "类型消息..."
	}
}

// 初始化所有已有任务
func initAllTasks() {
	cleanIllegalTasks()
	var tasks []RemindTask
	proxy.GetDB().Find(&tasks)
	for _, task := range tasks {
		s, err := task.genSchedule()
		if err != nil {
			continue
		}
		err = task.addJob(s)
		if err != nil {
			log.Warnf("新增定时提醒任务失败：%v", err)
		}
	}
}
