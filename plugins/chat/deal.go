package chat

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Dealer func(ctx *zero.Ctx, question string) message.Message

var dealers = []Dealer{ // 在此添加新的Dealer即可，其它事宜会自动处理
	WhoAreYou,
	IDoNotKnow,
}

func dealChat(ctx *zero.Ctx) {
	question := ctx.ExtractPlainText()
	// 优先尝试自定义问答
	msg := DIYDialogue(ctx, question)
	if len(msg) > 0 {
		ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, msg...)...)
		return
	}
	// 自定义问答无内容，则仅处理OnlyToMe且非空消息
	if !ctx.Event.IsToMe || len(question) == 0 {
		return
	}
	for _, deal := range dealers {
		msg = deal(ctx, question)
		if len(msg) > 0 {
			ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, msg...)...)
			return
		}
	}
}

// DIYDialogue Dealer: 用户自定义对话
func DIYDialogue(ctx *zero.Ctx, question string) message.Message {
	if !ctx.Event.IsToMe && proxy.GetConfigBool("onlytome") {
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

// IDoNotKnow Dealer: XX不知道
func IDoNotKnow(ctx *zero.Ctx, question string) message.Message {
	return message.Message{message.Text(fmt.Sprintf("%v不知道哦", utils.GetBotNickname()))}
}
