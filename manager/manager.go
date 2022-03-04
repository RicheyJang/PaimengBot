package manager

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/RicheyJang/PaimengBot/utils/rules"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	levelopt "github.com/syndtr/goleveldb/leveldb/opt"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type PluginHook func(condition *PluginCondition, ctx *zero.Ctx) error
type FileHook func(event fsnotify.Event) error

// PluginManager 插件管理器结构
type PluginManager struct {
	engine   *zero.Engine // zeroBot引擎
	configs  *viper.Viper // viper配置实例
	db       *gorm.DB     // DB
	dbConfig DBConfig     // 数据库配置
	leveldb  *leveldb.DB  // LevelDB实例

	plugins     map[string]*PluginProxy // plugin.key -> pluginContext
	preHooks    []PluginHook            // 插件Pre Hook
	postHooks   []PluginHook            // 插件Post Hook
	configHooks []FileHook              // 配置文件更改 Hook
}

// NewPluginManager 新建插件管理器
func NewPluginManager() *PluginManager {
	m := &PluginManager{
		engine:  zero.New(),
		configs: viper.New(),
		plugins: make(map[string]*PluginProxy),
	}
	// 添加前置Pre Hook
	m.engine.UsePreHandler(rules.SkipGuildMessage) // TODO 暂时忽略所有频道消息，原因：ZeroBot无法正常发送频道消息
	m.engine.UsePreHandler(m.preHandlerWithHook)
	// 添加后置Post Hook
	m.engine.UsePostHandler(m.postHandlerWithHook)
	return m
}

// RegisterPlugin 注册一个插件，并返回插件代理，用于添加事件动作、读写配置、获取插件锁、添加定时任务
func (manager *PluginManager) RegisterPlugin(info PluginInfo) *PluginProxy {
	thisPkgName := utils.GetPkgNameByFunc(NewPluginManager)
	key := utils.CallerPackageName(thisPkgName) // 暂时使用包名作为key值
	// 注册插件
	if len(info.Name) == 0 { // 无名插件
		log.Errorf("插件注册失败：<%s>没有设置Name", key)
		return nil
	}
	if _, ok := manager.plugins[key]; ok { // 已存在同名插件
		log.Errorf("插件注册失败：已存在同名插件%s", key)
		return nil
	}
	proxy := &PluginProxy{ // 创建插件代理
		key: key,
		u:   manager,
		c: PluginCondition{
			Key:        key,
			PluginInfo: info,
		},
	}
	manager.plugins[key] = proxy
	log.Infof("成功注册插件：%s", proxy.key)
	// 返回上下文
	return proxy
}

// FlushConfig 从文件中刷新所有插件配置，若文件不存在将会把配置写入该文件
func (manager *PluginManager) FlushConfig(configPath string, configFileName string) error {
	manager.configs.AddConfigPath(configPath)
	manager.configs.SetConfigFile(configFileName)
	fullPath := filepath.Join(configPath, configFileName)
	//fileType := filepath.Ext(fullPath)
	//manager.configs.SetConfigType(fileType)
	if utils.FileExists(fullPath) { // 配置文件已存在：合并自配置文件后重新写入
		err := manager.configs.MergeInConfig()
		if err != nil {
			log.Error("FlushConfig error in ReadInConfig err: ", err)
			return err
		}
		_ = manager.configs.WriteConfigAs(fullPath)
	} else { // 配置文件不存在：写入配置
		err := manager.configs.SafeWriteConfigAs(fullPath)
		if err != nil {
			log.Error("FlushConfig error in SafeWriteConfig err: ", err)
			return err
		}
	}
	manager.callAllConfigChangeHooks(fsnotify.Event{
		Name: fullPath,
		Op:   fsnotify.Create,
	})
	manager.configs.WatchConfig()
	manager.configs.OnConfigChange(func(in fsnotify.Event) {
		manager.callAllConfigChangeHooks(in)
		log.Infof("reload plugins config from %v", in.Name)
	})
	return nil
}

func (manager *PluginManager) callAllConfigChangeHooks(in fsnotify.Event) {
	manager.FlushAllAdminLevelFromConfig()     // 单独调用
	for _, hook := range manager.configHooks { // 执行配置文件更改时的各个Hook
		err := hook(in)
		if err != nil {
			log.Errorf("处理配置文件(%v)变更时出错：%v", in.Name, err)
		}
	}
}

// FlushAllAdminLevelFromConfig 从插件配置文件中刷新所有插件管理员权限等级，配置文件 插件名.adminlevel 配置项优先级高于代码预设info.AdminLevel
func (manager *PluginManager) FlushAllAdminLevelFromConfig() {
	plugins := GetAllPluginConditions()
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		// 获取配置文件中配置的管理员权限等级
		levelI := manager.getConfig(plugin.Key, consts.PluginConfigAdminLevelKey)
		if levelI == nil {
			continue
		}
		level := cast.ToInt(levelI)
		plugin.AdminLevel = level // 重设管理员权限等级
		if plugin.AdminLevel == 0 {
			log.Infof("依据配置文件，清除%v插件的管理员权限等级，非群管理员也可使用", plugin.Key)
		} else {
			log.Infof("依据配置文件，重设%v插件的管理员权限等级为%v", plugin.Key, plugin.AdminLevel)
		}
	}
}

func (manager *PluginManager) SetupDatabase(config DBConfig) error {
	// 1. 初始化关系型数据库
	gormC := &gorm.Config{ // 1.1 数据库配置
		Logger: utils.NewGormLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_", // 表名前缀，`User`表为`t_users`
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	}
	config.Type = strings.ToLower(config.Type)
	manager.dbConfig = config
	// 1.2 连接数据库
	switch config.Type {
	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host, config.Port, config.Name)
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:                       dsn,   // DSN data source name
			DefaultStringSize:         256,   // string 类型字段的默认长度
			SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
		}), gormC)
		if err != nil {
			log.Errorf("初始化数据库失败；%v", err)
			return err
		}
		manager.db = db
		log.Infof("初始化MySQL数据库成功：%v", dsn)
	case PostgreSQL:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.User, config.Passwd, config.Name, config.Port)
		db, err := gorm.Open(postgres.Open(dsn), gormC)
		if err != nil {
			log.Errorf("初始化数据库失败；%v", err)
			return err
		}
		manager.db = db
		log.Infof("初始化Postgresql数据库成功：%v", dsn)
	case SQLite:
		dsn := config.Name
		// 创建文件夹
		prePath, _ := filepath.Split(dsn)
		if len(prePath) > 0 {
			_, err := utils.MakeDirWithMode(prePath, 0o755)
			if err != nil {
				log.Errorf("初始化创建SQLite数据库文件夹失败；%v", err)
				return err
			}
		}
		// 连接数据库
		db, err := gorm.Open(sqlite.Open(dsn), gormC)
		if err != nil {
			log.Errorf("初始化数据库失败；%v", err)
			return err
		}
		manager.db = db
		log.Infof("初始化SQLite数据库成功：%v", dsn)
	default:
		return errors.New("暂不支持此类型数据库")
	}
	// 2. 初始化K-V数据库
	levelDB, err := leveldb.OpenFile(consts.DefaultLevelDBDir, &levelopt.Options{
		WriteBuffer: 128 * levelopt.KiB,
	})
	if err != nil {
		log.Errorf("初始化GoLevelDB失败，err: %v", err)
		return err
	}
	manager.leveldb = levelDB
	log.Infof("初始化K-V数据库成功：goleveldb")
	return nil
}

// GetAllPluginConditions 获取所有插件的详细信息
func (manager *PluginManager) GetAllPluginConditions() []*PluginCondition {
	var res []*PluginCondition
	for _, c := range manager.plugins {
		if c == nil {
			continue
		}
		res = append(res, &c.c)
	}
	return res
}

// GetPluginConditionByKey 按Key获取插件的详细信息
func (manager *PluginManager) GetPluginConditionByKey(key string) *PluginCondition {
	if p, ok := manager.plugins[key]; ok {
		return &p.c
	}
	return nil
}

// AddPreHook 添加前置hook
func (manager *PluginManager) AddPreHook(hook ...PluginHook) {
	manager.preHooks = append(manager.preHooks, hook...)
}

// AddPostHook 添加后置hook
func (manager *PluginManager) AddPostHook(hook ...PluginHook) {
	manager.postHooks = append(manager.postHooks, hook...)
}

// WhenConfigFileChange 添加配置文件变更时的hook
func (manager *PluginManager) WhenConfigFileChange(hook ...FileHook) {
	manager.configHooks = append(manager.configHooks, hook...)
}

// GetDB 获取数据库
func (manager *PluginManager) GetDB() *gorm.DB {
	return manager.db
}

// GetLevelDB 获取LevelDB: 一个K-V数据库
func (manager *PluginManager) GetLevelDB() *leveldb.DB {
	return manager.leveldb
}

// ---- 非公开方法 ----

// 默认插件管理器
var defaultManager = NewPluginManager()

// 通过Matcher获取其从属的PluginProxy
func (manager *PluginManager) getProxyByMatcher(matcher *zero.Matcher) *PluginProxy {
	if manager == nil {
		return nil
	}
	/* TODO　由于zeroBot在处理event时会copy原matcher，无法使用原matcher进行映射
	   因此只能通过matcher.Handler来拿取其所在的包名，来作为插件的key */
	key := utils.GetPkgNameByFunc(matcher.Handler)
	res, ok := manager.plugins[key]
	if !ok {
		log.Warnf("getProxyByMatcher No Such handler pkg name as key=%s", key)
		return nil
	}
	return res
}

// 前置总Hook Handler，会调用所有前置hook
func (manager *PluginManager) preHandlerWithHook(ctx *zero.Ctx) bool {
	matcher := ctx.GetMatcher()
	if matcher == nil {
		log.Debug("preHookLogMatcher matcher == nil")
		return true
	}
	proxy := manager.getProxyByMatcher(matcher)
	if proxy == nil { // 未注册插件
		log.Debug("preHookLogMatcher proxy == nil")
		return true
	}
	log.Infof("[Start] 事件即将被 <%s> 插件处理", proxy.key)
	// 调用所有前置hook
	for _, hook := range manager.preHooks {
		err := hook(&proxy.c, ctx)
		if err != nil {
			log.Infof("[End] <%s> 插件处理被 pre hook 取消，原因: %v", proxy.key, err)
			panic(consts.AbortLogIgnoreSymbol + err.Error()) // TODO 由于暂时没有Abort机制，只能使用panic来阻断执行
		}
	}
	log.Infof("[Begin] 前置Hook检查完毕，正式开始被 <%s> 插件处理", proxy.key)
	return true
}

// 后置总Hook Handler，会调用所有后置hook
func (manager *PluginManager) postHandlerWithHook(ctx *zero.Ctx) {
	matcher := ctx.GetMatcher()
	if matcher == nil {
		log.Debug("postHookLogMatcher matcher == nil")
		return
	}
	proxy := manager.getProxyByMatcher(matcher)
	if proxy == nil {
		log.Debug("postHookLogMatcher proxy == nil")
		return
	}
	log.Infof("[End] 事件被 <%s> 插件处理完毕\n", proxy.key)
	// 调用所有后置hook
	for _, hook := range manager.postHooks {
		err := hook(&proxy.c, ctx)
		if err != nil {
			log.Infof("[Tip] other matchers has been blocked by post hook reason: %v", err)
			panic(consts.AbortLogIgnoreSymbol + err.Error()) // TODO 由于暂时没有Abort机制，只能使用panic来阻断执行
		}
	}
}

// 添加配置并设置默认值
func (manager *PluginManager) addConfig(prefix string, key string, defaultValue interface{}) {
	if len(prefix) > 0 {
		key = fmt.Sprintf("%s.%s", prefix, key)
	}
	manager.configs.SetDefault(key, defaultValue)
}

// 获取配置
func (manager *PluginManager) getConfig(prefix string, key string) interface{} {
	if len(prefix) > 0 {
		key = fmt.Sprintf("%s.%s", prefix, key)
	}
	return manager.configs.Get(key)
}
