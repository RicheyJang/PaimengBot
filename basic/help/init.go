package help

import (
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "帮助",
	Usage: `
用法：
	帮助：展示已有的所有功能
	帮助[插件名或命令]：展示具体某个功能的详细帮助
`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"帮助", "help", "功能"}, zero.OnlyToMe).SetBlock(true).SetPriority(5).Handle(helpHandle)
}

func helpHandle(ctx *zero.Ctx) {
	isSuper := utils.IsSuperUser(ctx.Event.UserID)
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if len(arg) == 0 {
		level := auth.GetGroupUserPriority(ctx.Event.GroupID, ctx.Event.UserID)
		ctx.SendChain(formSummaryHelpMsg(isSuper, level))
	} else {
		ctx.SendChain(formSingleHelpMsg(arg, isSuper))
	}
}
