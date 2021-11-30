package ban

import (
	"fmt"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm/clause"
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
	if ctx.Event == nil {
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

// SetUserPluginStatus 设置指定用户的指定插件状态（于数据库）
func SetUserPluginStatus(status bool, userID int64, plugin *manager.PluginCondition, period time.Duration) error {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else { // plugin 为空，代表所有插件
		key = AllPluginKey
	}
	// 更新数据库
	var preUser dao.UserSetting
	proxy.GetDB().Take(&preUser, userID)
	preUser.ID = userID
	if status { // 启用
		preUser.BlackPlugins = delPluginKey(preUser.BlackPlugins, key)
	} else { // 关闭
		preUser.BlackPlugins = addPluginKey(preUser.BlackPlugins, key)
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"black_plugins"}), // Upsert
	}).Create(&preUser).Error; err != nil {
		log.Errorf("set user(%v) black_plugins error(sql): %v", userID, err)
		return err
	}
	log.Infof("设置用户%v插件<%v>状态：%v", userID, key, status)
	// 定时事件
	if period > 0 {
		_, _ = proxy.AddScheduleOnceFunc(period, func() {
			_ = SetUserPluginStatus(!status, userID, plugin, 0)
		})
	}
	return nil
}

// SetGroupPluginStatus 设置指定群的指定插件状态（于数据库）
func SetGroupPluginStatus(status bool, groupID int64, plugin *manager.PluginCondition, period time.Duration) error {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else { // plugin 为空，代表所有插件
		key = AllPluginKey
	}
	// 更新数据库
	var preGroup dao.GroupSetting
	proxy.GetDB().Take(&preGroup, groupID)
	preGroup.ID = groupID
	if status { // 启用
		preGroup.BlackPlugins = delPluginKey(preGroup.BlackPlugins, key)
	} else { // 关闭
		preGroup.BlackPlugins = addPluginKey(preGroup.BlackPlugins, key)
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"black_plugins"}), // Upsert
	}).Create(&preGroup).Error; err != nil {
		log.Errorf("set group(%v) black_plugins error(sql): %v", groupID, err)
		return err
	}
	log.Infof("设置群%v插件<%v>状态：%v", groupID, key, status)
	// 定时事件
	if period > 0 {
		_, _ = proxy.AddScheduleOnceFunc(period, func() {
			_ = SetGroupPluginStatus(!status, groupID, plugin, 0)
		})
	}
	return nil
}

// GetUserPluginStatus 获取用户插件状态（能否使用）
func GetUserPluginStatus(userID int64, plugin *manager.PluginCondition) bool {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else { // plugin 为空，代表所有插件
		key = AllPluginKey
	}
	// 查询
	var preUser dao.UserSetting
	proxy.GetDB().Take(&preUser, userID)
	return !(hasPluginKey(preUser.BlackPlugins, key) || hasPluginKey(preUser.BlackPlugins, AllPluginKey))
}

// GetGroupPluginStatus 获取群插件状态（能否使用）
func GetGroupPluginStatus(groupID int64, plugin *manager.PluginCondition) bool {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else { // plugin 为空，代表所有插件
		key = AllPluginKey
	}
	// 查询
	var preGroup dao.GroupSetting
	proxy.GetDB().Take(&preGroup, groupID)
	return !(hasPluginKey(preGroup.BlackPlugins, key) || hasPluginKey(preGroup.BlackPlugins, AllPluginKey))
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
