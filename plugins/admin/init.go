package admin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/rules"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name:     "群管理",
	Classify: "群功能",
	Usage: `用法：（需要设为管理员）
	踢了 [QQ号或@]：并在后续询问中答复"是"，则将指定QQ号或@的人踢出本群
	禁言 [QQ号或@] [时长]：将指定QQ号或@的人指定时长（时长为0则解除禁言）
另外，回复某条消息"禁言 [时长]"，则可以将原消息发送者禁言指定时长
示例：
	禁言 123456 30m：将123456禁言30分钟
	禁言 @XXX 1d12h：将XXX禁言1天12小时
	禁言 123456 0：将123456解除禁言`,
	AdminLevel: 5,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"踢了"}, zero.OnlyGroup).SetBlock(true).ThirdPriority().Handle(kickSomeone)
	proxy.OnCommands([]string{"禁言"}, zero.OnlyGroup).SetBlock(true).ThirdPriority().Handle(muteSomeone)
	proxy.OnMessage(zero.OnlyGroup, rules.ReplyAndCommands("禁言")).SetBlock(true).ThirdPriority().Handle(muteReply)
}

func kickSomeone(ctx *zero.Ctx) {
	id, _, err := analysisArgs(ctx, false)
	if err != nil {
		log.Errorf("解析参数错误：%v", err)
		return
	}
	if id <= 0 {
		ctx.Send("请指定踢出的QQ号或@")
		return
	}
	// 获取确认
	ctx.Send(fmt.Sprintf("确认踢出用户%v？", id))
	event := utils.WaitNextMessage(ctx)
	if event == nil { // 无回应
		return
	}
	confirm := strings.TrimSpace(event.Message.ExtractPlainText())
	if confirm == "是" || confirm == "确定" || confirm == "确认" {
		// 确认踢出
		ctx.SetGroupKick(ctx.Event.GroupID, id, false)
		log.Infof("将用户%v踢出群%v", id, ctx.Event.GroupID)
	} else {
		ctx.Send("已取消")
	}
}

func muteSomeone(ctx *zero.Ctx) {
	id, duration, err := analysisArgs(ctx, true)
	if err != nil {
		log.Errorf("解析参数错误：%v", err)
		return
	}
	if id < 0 {
		ctx.Send("请指定禁言的QQ号或@")
		return
	}
	if duration != 0 && duration <= time.Minute { // 至少禁言1分钟
		duration = time.Minute
	}
	// 禁言
	if id == 0 { // WARNING Go-Cqhttp的全员禁言目前无法使用，会在下一版本修复
		ctx.SetGroupWholeBan(ctx.Event.GroupID, duration != 0)
	} else {
		ctx.SetGroupBan(ctx.Event.GroupID, id, int64(duration/time.Second))
	}
	// log
	if duration == 0 {
		log.Infof("将用户%v解除禁言", id)
	} else {
		log.Infof("将用户%v禁言%v", id, duration)
	}
}

func muteReply(ctx *zero.Ctx) {
	replyID := cast.ToInt64(ctx.State["reply_id"])
	// 追溯消息
	msg := ctx.GetMessage(replyID)
	if msg.Sender == nil {
		ctx.Send("消息失效了")
		return
	}
	// 解析时长
	args := strings.TrimSpace(utils.GetArgs(ctx))
	var duration time.Duration
	if args == "0" {
		duration = 0
	} else {
		seconds, err := parseDurationWithDay(args)
		if err != nil {
			ctx.Send("时长格式错误")
			return
		}
		duration = seconds
	}
	if duration != 0 && duration <= time.Minute { // 至少禁言1分钟
		duration = time.Minute
	}
	// 禁言
	// WARNING Go-Cqhttp并没有提供AnonymousFlag字段，因此无法支持禁言匿名用户
	ctx.SetGroupBan(ctx.Event.GroupID, msg.Sender.ID, int64(duration/time.Second))
	log.Infof("将用户%v禁言%s", msg.Sender.ID, duration)
}

func analysisArgs(ctx *zero.Ctx, parseTime bool) (ID int64, seconds time.Duration, err error) {
	parseID := true
	// @
	for _, msg := range ctx.Event.Message {
		if msg.Type == "at" {
			parseID = false
			ID, _ = strconv.ParseInt(msg.Data["qq"], 10, 64)
			if !parseTime {
				return
			}
		}
	}
	// 文字
	args := strings.TrimSpace(utils.GetArgs(ctx))
	subs := strings.Split(args, " ")
	if len(subs) < 1 { // 要么要有QQ号，要么要有时长
		ctx.Send("参数不足")
		return 0, 0, fmt.Errorf("参数不足")
	}
	// 解析QQ号
	if parseID {
		ID, err = strconv.ParseInt(subs[0], 10, 64)
		if err != nil {
			ctx.Send("QQ号格式错误")
			return
		}
	}
	// 解析时长
	if parseTime {
		if subs[len(subs)-1] == "0" {
			seconds = 0
			return
		}
		seconds, err = parseDurationWithDay(subs[len(subs)-1])
		if err != nil {
			ctx.Send("时长格式错误")
			return
		}
	}
	return
}

func parseDurationWithDay(duration string) (res time.Duration, err error) {
	if len(duration) == 0 {
		return 0, fmt.Errorf("duration too short")
	}
	if match, _ := regexp.MatchString("^[1-9]+d.*", duration); match {
		dIndex := strings.IndexRune(duration, 'd')
		day, _ := strconv.Atoi(duration[:dIndex])
		res = 24 * time.Hour * time.Duration(day)
		duration = duration[dIndex+1:]
		if len(duration) == 0 {
			return
		}
	}
	other, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return other + res, err
}
