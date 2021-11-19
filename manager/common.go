package manager

import (
	"github.com/RicheyJang/PaimengBot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ---- 初始化相关 ----

// 全局初始化
func init() {
	// 初始化命令行参数、主配置、logger
	PaimengBot.DoPreWorks()

	// 读取插件配置
	err := FlushConfig(".", "config-plugins.yaml")
	if err != nil {
		log.Fatal("FlushConfig err: ", err)
	}
	// 初始化数据库
	dbV := viper.Sub("db")
	dbC := new(DBConfig)
	err = dbV.Unmarshal(dbC)
	if err != nil {
		log.Fatal("Unmarshal DB Config err: ", err)
	}
	err = SetupDatabase(dbV.GetString("type"), *dbC)
	if err != nil {
		log.Fatal("SetupDatabase err: ", err)
	}
}

// FlushConfig 从文件中刷新所有插件配置
func FlushConfig(configPath, configFileName string) error {
	return defaultManager.FlushConfig(configPath, configFileName)
}

// DBConfig 数据库设置
type DBConfig struct {
	Host   string // 地址
	Port   int    // 端口
	User   string // 用户名
	Passwd string // 密码
	Name   string // 数据库名
}

// SetupDatabase 初始化数据库
func SetupDatabase(tp string, config DBConfig) error {
	return defaultManager.SetupDatabase(tp, config)
}

// ---- 插件相关 ----

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
