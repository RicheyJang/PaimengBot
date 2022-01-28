package chat

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

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
	question := ctx.ExtractPlainText()
	// 优先尝试自定义问答
	msg := DIYDialogue(ctx, question)
	if len(msg) > 0 {
		sendChatMessage(ctx, msg)
		return
	}
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

func sendChatMessage(ctx *zero.Ctx, msg message.Message) {
	if utils.IsMessagePrimary(ctx) || !proxy.GetConfigBool("at") { // 私聊或无需@
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
	if !ctx.Event.IsToMe && proxy.GetConfigBool("onlytome") { // 若配置了onlytome，则仅处理onlytome消息
		return nil
	}
	if utils.IsMessageGroup(ctx) {
		msg := GetDialogue(ctx.Event.GroupID, question)
		if len(msg) > 0 {
			return msg
		}
	}
	return GetDialogue(0, question)
}

// WhoAreYou Dealer: 自我介绍
func WhoAreYou(ctx *zero.Ctx, question string) message.Message {
	if question == "你是谁" || question == "是谁" || question == "你是什么" || question == "是什么" {
		return message.Message{message.Text(proxy.GetConfigString("default.self"))}
	}
	return nil
}

// PluginName Dealer: 问题为插件名，返回帮助信息
func PluginName(ctx *zero.Ctx, question string) message.Message {
	plugins := manager.GetAllPluginConditions()
	for _, plugin := range plugins {
		if question == plugin.Name {
			return message.Message{message.Text(
				fmt.Sprintf("这是%v的一个功能名哟，想知道这个功能怎么使用的话，请说：\n帮助 %v", utils.GetBotNickname(), question))}
		}
	}
	return nil
}

// IDoNotKnow Dealer: XX不知道
func IDoNotKnow(ctx *zero.Ctx, question string) message.Message {
	return message.Message{message.Text(fmt.Sprintf("%v不知道哦", utils.GetBotNickname()))}
}
