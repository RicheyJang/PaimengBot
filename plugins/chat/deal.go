package chat

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Dealer func(ctx *zero.Ctx, question string) message.Message

var dealers = []Dealer{
	DIYDialogue,
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

func DIYDialogue(ctx *zero.Ctx, question string) message.Message {
	if utils.IsMessageGroup(ctx) {
		msg := GetDialogue(ctx.Event.GroupID, question)
		if len(msg) > 0 {
			return msg
		}
	}
	return GetDialogue(0, question)
}

func IDoNotKnow(ctx *zero.Ctx, question string) message.Message {
	return message.Message{message.Text(fmt.Sprintf("%v不知道哦", utils.GetBotNickname()))}
}
