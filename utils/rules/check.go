package rules

import (
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// CheckDetailType Rule:检查事件DetailType(MessageType/NoticeType/RequestType)
func CheckDetailType(tp string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if ctx.Event != nil {
			return ctx.Event.DetailType == tp
		}
		return false
	}
}

// SkipGroupAnonymous Rule:不处理群匿名消息
func SkipGroupAnonymous(ctx *zero.Ctx) bool {
	return !utils.IsGroupAnonymous(ctx)
}

// SkipGuildMessage Rule:不处理频道消息事件
func SkipGuildMessage(ctx *zero.Ctx) bool {
	return !utils.IsMessageGuild(ctx)
}
