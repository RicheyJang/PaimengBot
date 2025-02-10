package ban

import (
	"time"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

// SetUserPluginStatus 设置指定用户的指定插件状态（于数据库）
func SetUserPluginStatus(status bool, userID int64, plugin *manager.PluginCondition, period time.Duration) error {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else {
		panic("SetGroupPluginStatus: plugin is nil")
	}
	// 更新数据库
	var preUser dao.UserSetting
	proxy.GetDB().Take(&preUser, userID)
	preUser.ID = userID
	if preUser.WhiteMode {
		if status { // 启用
			preUser.WhitePlugins = addPluginKey(preUser.WhitePlugins, key)
		} else { // 关闭
			preUser.WhitePlugins = delPluginKey(preUser.WhitePlugins, key)
		}
	} else {
		if status { // 启用
			preUser.BlackPlugins = delPluginKey(preUser.BlackPlugins, key)
		} else { // 关闭
			preUser.BlackPlugins = addPluginKey(preUser.BlackPlugins, key)
		}
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"black_plugins", "white_plugins"}), // Upsert
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
	} else {
		panic("SetGroupPluginStatus: plugin is nil")
	}
	// 更新数据库
	var preGroup dao.GroupSetting
	proxy.GetDB().Take(&preGroup, groupID)
	preGroup.ID = groupID
	if preGroup.WhiteMode {
		if status { // 启用
			preGroup.WhitePlugins = addPluginKey(preGroup.WhitePlugins, key)
		} else { // 关闭
			preGroup.WhitePlugins = delPluginKey(preGroup.WhitePlugins, key)
		}
	} else {
		if status { // 启用
			preGroup.BlackPlugins = delPluginKey(preGroup.BlackPlugins, key)
		} else { // 关闭
			preGroup.BlackPlugins = addPluginKey(preGroup.BlackPlugins, key)
		}
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"black_plugins", "white_plugins"}), // Upsert
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
	if preUser.WhiteMode { // 白名单模式
		return hasPluginKey(preUser.WhitePlugins, key)
	} else {
		return !(hasPluginKey(preUser.BlackPlugins, key) || hasPluginKey(preUser.BlackPlugins, AllPluginKey))
	}
}

// GetGroupPluginStatus 获取群插件状态（能否使用）
func GetGroupPluginStatus(groupID int64, plugin *manager.PluginCondition) bool {
	// 获取插件Key
	var key string
	if plugin != nil {
		key = plugin.Key
	} else {
		key = AllPluginKey
	}
	// 查询
	var preGroup dao.GroupSetting
	proxy.GetDB().Take(&preGroup, groupID)
	if preGroup.WhiteMode { // 白名单模式
		return hasPluginKey(preGroup.WhitePlugins, key)
	} else {
		return !(hasPluginKey(preGroup.BlackPlugins, key) || hasPluginKey(preGroup.BlackPlugins, AllPluginKey))
	}
}

// GetGroupRunningMode 获取群运行模式（白名单模式为true，黑名单为false）
func GetGroupRunningMode(groupID int64) bool {
	var preGroup dao.GroupSetting
	proxy.GetDB().Take(&preGroup, groupID)
	return preGroup.WhiteMode
}

// GetUserRunningMode 获取用户运行模式（白名单模式为true，黑名单为false）
func GetUserRunningMode(userID int64) bool {
	var preUser dao.UserSetting
	proxy.GetDB().Take(&preUser, userID)
	return preUser.WhiteMode
}
