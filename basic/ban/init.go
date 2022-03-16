package ban

import (
	"fmt"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "功能开关",
	Usage: `
用法：
	*只有本群最高权限群管理员在群聊中才可触发*
	开启\关闭[功能] [时长]?：将开启\关闭本群的指定功能，时长为可选项，形式参照示例
	封禁[用户ID] [功能]? [时长]?：封禁指定用户使用指定功能（当指定功能时）或全部功能，时长为可选项，形式参照示例
	解封[用户ID] [功能]?：解封指定用户使用指定功能，时长为可选项，形式参照示例
	黑名单：获取所有被封禁用户的被封禁功能列表
示例：
	封禁123456：封禁用户ID为123456的所有功能
	封禁123456 25m：封禁用户ID为123456的所有功能25分钟
	封禁123456 翻译 1h30m：封禁用户ID123456的翻译功能1小时零30分钟
`,
	SuperUsage: `
用法：
	在私聊中：
		使用开启\关闭[功能] [时长]?命令，将针对所有用户和群开启\关闭该功能（全局Ban）
		还可通过 开启\关闭[群ID] [功能] [时长]? 来开启\关闭指定群的指定功能
		黑名单：获取所有被封禁用户、群的被封禁功能列表
	在群聊中，等同于最高权限群管理员执行命令
`,
	AdminLevel: 1,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"开启"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(openPlugin)
	proxy.OnCommands([]string{"关闭"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(closePlugin)
	proxy.OnCommands([]string{"封禁", "ban", "Ban"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(banUser)
	proxy.OnCommands([]string{"解封", "unban", "Unban"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(unbanUser)
	proxy.OnCommands([]string{"黑名单"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(showBlack)
	manager.AddPreHook(checkPluginStatus)
}

const AllPluginKey = "all"

func checkPluginStatus(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	if !utils.IsMessage(ctx) { //仅处理消息类型事件
		return nil
	}
	// 群ban
	if ctx.Event.GroupID != 0 && !GetGroupPluginStatus(ctx.Event.GroupID, condition) {
		return fmt.Errorf("此插件<%v>在此群(%v)已被关闭", condition.Key, ctx.Event.GroupID)
	}
	// 个人ban
	if ctx.Event.UserID != 0 && !GetUserPluginStatus(ctx.Event.UserID, condition) {
		return fmt.Errorf("此插件<%v>对此用户(%v)已被禁用", condition.Key, ctx.Event.UserID)
	}
	// 全局ban
	if !GetUserPluginStatus(0, condition) {
		return fmt.Errorf("此插件<%v>已全局禁用", condition.Key)
	}
	return nil
}

func dealUserPluginStatus(ctx *zero.Ctx, status bool, userID int64, plugin *manager.PluginCondition, period time.Duration) {
	if status == GetUserPluginStatus(userID, plugin) {
		ctx.Send("请不要重复开关功能哦")
		return
	}
	err := SetUserPluginStatus(status, userID, plugin, period)
	if err != nil {
		ctx.Send("失败了...")
	} else {
		ctx.Send("好哒")
	}
}

func dealGroupPluginStatus(ctx *zero.Ctx, status bool, groupID int64, plugin *manager.PluginCondition, period time.Duration) {
	if status == GetGroupPluginStatus(groupID, plugin) {
		ctx.Send("请不要重复开关功能哦")
		return
	}
	err := SetGroupPluginStatus(status, groupID, plugin, period)
	if err != nil {
		ctx.Send("失败了...")
	} else {
		ctx.Send("好哒")
	}
}

func hasPluginKey(org, key string) bool {
	name := fmt.Sprintf("|%s|", key)
	return strings.Contains(org, name)
}

func addPluginKey(org, key string) string {
	if len(org) == 0 || !strings.HasSuffix(org, "|") {
		org += "|"
	}
	if !hasPluginKey(org, key) {
		return org + key + "|"
	}
	return org
}

func delPluginKey(org, key string) string {
	name := fmt.Sprintf("|%s|", key)
	return strings.ReplaceAll(org, name, "|")
}

// 通过插件名或Key查找插件Condition
func findPluginByName(name string) *manager.PluginCondition {
	plugins := manager.GetAllPluginConditions()
	for _, plugin := range plugins {
		if plugin.Name == name || plugin.Key == name {
			return plugin
		}
	}
	return nil
}
