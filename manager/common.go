package manager

import (
	"github.com/RicheyJang/PaimengBot/utils"
)

// 全局初始化
func init() {
	utils.DoPreWorks()
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string // Need 插件名称
	Usage       string // Need 插件用法描述
	SuperUsage  string // Option 插件超级用户用法描述
	Classify    string // Option 插件分类，为空时代表默认分类
	IsHidden    bool   // Option 是否为隐藏插件
	IsSuperOnly bool   // Option 是否为超级用户专属插件
	AdminLevel  int    // Option 群管理员使用最低级别： 0 表示非群管理员专用插件 >0 表示数字越低，权限要求越高
}

// FlushConfig 从文件中刷新所有插件配置
func FlushConfig(configPath, configFileName string) error {
	return defaultManager.FlushConfig(configPath, configFileName)
}

// RegisterPlugin 注册一个插件至默认插件管理器，并返回插件代理
func RegisterPlugin(info PluginInfo) *PluginProxy {
	return defaultManager.RegisterPlugin(info)
}

// GetAllPluginConditions 获取所有插件的详细信息
func GetAllPluginConditions() []PluginCondition {
	return defaultManager.GetAllPluginConditions()
}

// AddPreHook 添加前置hook
func AddPreHook(hook ...PluginHook) {
	defaultManager.AddPreHook(hook...)
}

// AddPostHook 添加后置hook
func AddPostHook(hook ...PluginHook) {
	defaultManager.AddPostHook(hook...)
}
