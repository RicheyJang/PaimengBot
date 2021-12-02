package help

import (
	"math"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/basic/dao"
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
	level := auth.GetGroupUserPriority(ctx.Event.GroupID, ctx.Event.UserID)
	blacks := getBlackKeys(ctx.Event.UserID, ctx.Event.GroupID)
	if utils.IsGroupAnonymous(ctx) { // 匿名用户单独处理
		isSuper = false
		level = math.MaxInt
	}
	if len(arg) == 0 {
		ctx.SendChain(formSummaryHelpMsg(isSuper, utils.IsMessagePrimary(ctx), level, blacks))
	} else {
		ctx.SendChain(formSingleHelpMsg(arg, isSuper, utils.IsMessagePrimary(ctx), level, blacks))
	}
}

func checkPluginCouldShow(plugin *manager.PluginCondition, isSuper, isPrimary bool, priority int) bool {
	if plugin == nil {
		return false
	}
	if plugin.IsSuperOnly && (!isSuper || !isPrimary) { // 超级用户专属
		return false
	}
	if plugin.AdminLevel > 0 && (priority == 0 || priority > plugin.AdminLevel) { // 管理员权限
		return false
	}
	return true
}

func getBlackKeys(userID, groupID int64) map[string]struct{} {
	var users []dao.UserSetting
	var groupS dao.GroupSetting
	proxy.GetDB().Find(&users, []int64{0, userID})
	if groupID != 0 {
		proxy.GetDB().Find(&groupS, groupID)
	}
	var usersKey string
	for _, user := range users {
		usersKey += user.BlackPlugins
	}
	return utils.FormSetByStrings(strings.Split(groupS.BlackPlugins, "|"),
		strings.Split(usersKey, "|"))
}
