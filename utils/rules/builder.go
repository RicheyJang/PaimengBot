package rules

import (
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// ReplyAndCommands Rule生成器: 含回复且文本中包含指定命令
func ReplyAndCommands(commands ...string) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		if len(ctx.Event.Message) < 2 || ctx.Event.Message[0].Type != "reply" { // 回复消息
			return false
		}
		ctx.State["reply_id"] = ctx.Event.Message[0].Data["id"]
		for i, msg := range ctx.Event.Message {
			if msg.Type == "text" && checkTextMsgCommands(ctx, msg.Data["text"], i, commands...) {
				return true
			}
		}
		return false
	}
}

// 检查文本消息中是否包含指定命令
func checkTextMsgCommands(ctx *zero.Ctx, msg string, textIndex int, commands ...string) bool {
	if len(msg) == 0 {
		return false
	}
	msg = strings.TrimSpace(msg)
	// 去除昵称前缀
	if !utils.IsMessagePrimary(ctx) {
		for _, nick := range utils.GetBotConfig().NickName {
			if strings.HasPrefix(msg, nick) {
				msg = msg[len(nick):]
				msg = strings.TrimSpace(msg)
				ctx.Event.IsToMe = true
				break
			}
		}
	}
	// 检查命令前缀
	if !strings.HasPrefix(msg, utils.GetBotConfig().CommandPrefix) {
		return false
	}
	cmdMessage := msg[len(utils.GetBotConfig().CommandPrefix):]
	// 检查命令
	for _, command := range commands {
		if strings.HasPrefix(cmdMessage, command) {
			ctx.State["command"] = command
			arg := strings.TrimLeft(cmdMessage[len(command):], " ")
			if len(ctx.Event.Message) > textIndex+1 {
				arg += ctx.Event.Message[textIndex+1:].ExtractPlainText()
			}
			ctx.State["args"] = arg
			return true
		}
	}
	return false
}
