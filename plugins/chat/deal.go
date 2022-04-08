package chat

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Dealer func(ctx *zero.Ctx, question string) message.Message

var dealers = []Dealer{ // 在此添加新的Dealer即可，其它事宜会自动处理
	WhoAreYou,
	PluginName,
	IDoNotKnow,
}

func dealChat(ctx *zero.Ctx) {
	question := preprocessQuestion(ctx.MessageString())
	// 优先尝试自定义问答
	msg := DIYDialogue(ctx, question)
	if len(msg) > 0 {
		sendChatMessage(ctx, msg)
		return
	}
	defer func() { // 若并没有回复消息，则无需统计
		if len(msg) == 0 {
			utils.SetNotStatistic(ctx)
		}
	}()
	// 自定义问答无内容，则仅处理OnlyToMe且非空消息
	if !ctx.Event.IsToMe || len(question) == 0 {
		return
	}
	for _, deal := range dealers {
		msg = deal(ctx, question)
		if len(msg) > 0 {
			sendChatMessage(ctx, msg)
			return
		}
	}
}

var singleCQTypeSet = map[string]struct{}{
	"record": {}, "video": {}, "rps": {}, "dice": {}, "shake": {}, "poke": {}, "share": {},
	"contact": {}, "location": {}, "music": {}, "reply": {}, "forward": {}, "node": {},
}

func sendChatMessage(ctx *zero.Ctx, msg message.Message) {
	doNotAt := false
	// 检查是否包含不能@的消息类型
	for _, seg := range msg {
		if _, ok := singleCQTypeSet[seg.Type]; ok {
			doNotAt = true
			break
		}
	}
	if doNotAt || utils.IsMessagePrimary(ctx) || !proxy.GetConfigBool("at") { // 私聊或无需@
		ctx.Send(msg)
	} else {
		ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, msg...)...)
	}
}

// DIYDialogue Dealer: 用户自定义对话
func DIYDialogue(ctx *zero.Ctx, question string) message.Message {
	if len(question) == 0 {
		return nil
	}
	if utils.IsMessageGroup(ctx) {
		msg := GetDialogue(ctx, ctx.Event.GroupID, question)
		if len(msg) > 0 {
			return msg
		}
	}
	return GetDialogue(ctx, 0, question) // 全局问答
}

// WhoAreYou Dealer: 自我介绍
func WhoAreYou(ctx *zero.Ctx, question string) message.Message {
	if question == "你是谁" || question == "是谁" ||
		question == "你是什么" || question == "是什么" ||
		question == "自我介绍" {
		return message.Message{message.Text(proxy.GetConfigString("default.self"))}
	}
	return nil
}

// PluginName Dealer: 问题为插件名，返回帮助信息
func PluginName(ctx *zero.Ctx, question string) message.Message {
	plugins := manager.GetAllPluginConditions()
	for _, plugin := range plugins {
		if question == plugin.Name {
			str := fmt.Sprintf("这是%v的一个功能名哟，想知道这个功能怎么使用的话，请说：\n帮助 %v", utils.GetBotNickname(), question)
			if !utils.IsMessagePrimary(ctx) {
				str = fmt.Sprintf("这是%v的一个功能名哟，想知道这个功能怎么使用的话，请说：\n%[1]v帮助 %v", utils.GetBotNickname(), question)
			}
			return message.Message{message.Text(str)}
		}
	}
	return nil
}

// IDoNotKnow Dealer: XX不知道
func IDoNotKnow(ctx *zero.Ctx, question string) message.Message {
	var a []string
	c := proxy.GetConfig("default.donotknow")
	switch c.(type) {
	case string:
		a = []string{cast.ToString(c)}
	case []string, []interface{}:
		a = cast.ToStringSlice(c)
	}
	if len(a) == 0 {
		a = []string{"{nickname}不知道哦"}
	}
	str := a[rand.Intn(len(a))]
	str = strings.ReplaceAll(str, "{nickname}", utils.GetBotNickname())
	return message.Message{message.Text(str)}
}
