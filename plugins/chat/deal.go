package chat

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Dealer func(ctx *zero.Ctx, question string) message.Message

var dealers = []Dealer{ // 在此添加新的Dealer即可，其它事宜会自动处理
	DIYDialogue,
	WhoAreYou,
	IDoNotKnow,
}

func dealChat(ctx *zero.Ctx) {
	question := ctx.ExtractPlainText()
	if len(question) == 0 {
		ctx.Send("？")
		return
	}
	for _, deal := range dealers {
		msg := deal(ctx, question)
		if len(msg) > 0 {
			ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, msg...)...)
			return
		}
	}
}

// DIYDialogue Dealer: 用户自定义对话
func DIYDialogue(ctx *zero.Ctx, question string) message.Message {
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
