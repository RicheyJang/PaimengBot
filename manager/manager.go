package manager

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type PluginHook func(condition *PluginCondition, ctx *zero.Ctx) error

// PluginManager 插件管理器结构
type PluginManager struct {
	engine   *zero.Engine // zeroBot引擎
	configs  *viper.Viper // viper配置实例
	db       *gorm.DB     // DB
	dbConfig DBConfig     // 数据库配置

	plugins   map[string]*PluginProxy // plugin.key -> pluginContext
	preHooks  []PluginHook            // 插件Pre Hook
	postHooks []PluginHook            // 插件Post Hook
}

// NewPluginManager 新建插件管理器
func NewPluginManager() *PluginManager {
	m := &PluginManager{
		engine:  zero.New(),
		configs: viper.New(),
		plugins: make(map[string]*PluginProxy),
	}
	// 添加前置Pre Hook
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
	if utils.FileExists(fullPath) { // 配置文件已存在：读出配置
		err := manager.configs.ReadInConfig()
		if err != nil {
			log.Error("FlushConfig error in ReadInConfig")
			return err
		}
	} else { // 配置文件不存在：写入配置
		err := manager.configs.SafeWriteConfigAs(fullPath)
		if err != nil {
			log.Error("FlushConfig error in SafeWriteConfig")
			return err
		}
	}
	manager.configs.WatchConfig()
	return nil
}

func (manager *PluginManager) SetupDatabase(config DBConfig) error {
	// 初始化数据库配置
	gormC := &gorm.Config{ // 数据库配置
		Logger: utils.NewGormLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_", // 表名前缀，`User`表为`t_users`
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	}
	config.Type = strings.ToLower(config.Type)
	manager.dbConfig = config
	// 连接数据库
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
		prePath := "."
		for i, c := range dsn {
			if c == '/' || c == '\\' {
				prePath = dsn[:i]
				break
			}
		}
		_, err := utils.MakeDirWithMode(prePath, 0o755)
		if err != nil {
			log.Errorf("初始化创建SQLite数据库文件夹失败；%v", err)
			return err
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

// GetDB 获取数据库
func (manager *PluginManager) GetDB() *gorm.DB {
	return manager.db
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
	log.Infof("The message starts to be handled by the <%s> plugin", proxy.key)
	// 调用所有前置hook
	for _, hook := range manager.preHooks {
		err := hook(&proxy.c, ctx)
		if err != nil {
			log.Infof("<%s> handle has been canceled by pre hook reason: %v", proxy.key, err)
			panic(consts.AbortLogIgnoreSymbol + err.Error()) // TODO 由于暂时没有Abort机制，只能使用panic来阻断执行
		}
	}
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
	log.Infof("The message is handled finish by the <%s> plugin\n", proxy.key)
	// 调用所有后置hook
	for _, hook := range manager.postHooks {
		err := hook(&proxy.c, ctx)
		if err != nil {
			log.Infof("other matchers has been blocked by post hook reason: %v", err)
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
