package manager

import (
	"io/fs"

	"github.com/RicheyJang/PaimengBot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// ---- 初始化相关 ----

// 全局初始化
func init() {
	// 初始化命令行参数、主配置、logger
	PaimengBot.DoPreWorks()

	// 初始化数据库
	dbV := viper.Sub("db")
	dbC := new(DBConfig)
	err := dbV.Unmarshal(dbC)
	if err != nil {
		log.Fatal("读取数据库配置出错 err: ", err)
	}
	err = SetupDatabase(*dbC)
	if err != nil {
		log.Fatal("初始化数据库连接失败 err: ", err)
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
	Type   string // 数据库类型
}

const (
	MySQL      = "mysql"
	PostgreSQL = "postgresql"
	SQLite     = "sqlite"
)

// SetupDatabase 初始化数据库
func SetupDatabase(config DBConfig) error {
	return defaultManager.SetupDatabase(config)
}

// ---- 插件相关 ----

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string // Need 插件名称
	Usage       string // Need 插件用法描述
	SuperUsage  string // Option 插件超级用户用法描述
	Classify    string // Option 插件分类，为空时代表默认分类
	IsPassive   bool   // Option 是否为被动插件：在帮助中被标识为被动功能；
	IsSuperOnly bool   // Option 是否为超级用户专属插件：消息性事件会自动加上SuperOnly检查；在帮助中只有超级用户私聊可见；
	AdminLevel  int    // Option 群管理员使用最低级别： 0 表示非群管理员专用插件 >0 表示数字越低，权限要求越高；在帮助中进行标识；配置文件中 插件名.adminlevel 配置项优先级高于此项
}

// RegisterPlugin 注册一个插件至默认插件管理器，并返回插件代理
func RegisterPlugin(info PluginInfo) *PluginProxy {
	return defaultManager.RegisterPlugin(info)
}

// GetAllPluginConditions 获取所有插件的详细信息
func GetAllPluginConditions() []*PluginCondition {
	return defaultManager.GetAllPluginConditions()
}

// GetPluginConditionByKey 按Key获取插件的详细信息
func GetPluginConditionByKey(key string) *PluginCondition {
	return defaultManager.GetPluginConditionByKey(key)
}

// AddPreHook 添加前置hook
func AddPreHook(hook PluginHook) *HookMatcher {
	return defaultManager.AddPreHook(hook)
}

// AddPostHook 添加后置hook
func AddPostHook(hook PluginHook) *HookMatcher {
	return defaultManager.AddPostHook(hook)
}

// WhenConfigFileChange 增加配置文件变更时的处理函数
func WhenConfigFileChange(hook ...FileHook) {
	defaultManager.WhenConfigFileChange(hook...)
}

// GetDB 获取数据库
func GetDB() *gorm.DB {
	return defaultManager.GetDB()
}

// GetStaticFile 获取指定静态文件
func GetStaticFile(name string) (fs.File, error) {
	return PaimengBot.GetStaticFS().Open("static/" + name)
}

// ReadStaticFile 读取指定静态文件
func ReadStaticFile(name string) ([]byte, error) {
	return PaimengBot.GetStaticFS().ReadFile("static/" + name)
}
