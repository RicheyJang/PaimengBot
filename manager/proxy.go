package manager

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

// PluginProxy 插件代理，呈现给插件，用于添加事件动作、读写配置、获取插件锁、添加定时任务
// 插件在注册后，应只与此代理交互，与Manager再无交际
type PluginProxy struct {
	key string         // 插件Key
	u   *PluginManager // 所从属的插件管理器

	c PluginCondition // 插件状态（被管理器控制）
}

// ---- 事件动作 ----

// On 添加新的指定消息类型的匹配器
// zero.Engine.On的附加处理拷贝
func (p *PluginProxy) On(tp string, rules ...zero.Rule) *zero.Matcher {
	rules = p.checkRules(rules...)
	matcher := p.u.engine.On(tp, rules...)
	return matcher
}

// OnCommands 添加新的命令匹配器
func (p *PluginProxy) OnCommands(cmd []string, rules ...zero.Rule) *zero.Matcher {
	rules = p.checkRules(rules...)
	matcher := p.u.engine.OnCommandGroup(cmd, rules...)
	// 检查是否包含Rule：SuperUserPermission
	hasSuper := false
	for _, rule := range rules {
		if utils.IsSameFunc(rule, zero.SuperUserPermission) {
			hasSuper = true
			break
		}
	}
	// 添加命令记录
	if !hasSuper {
		p.c.NormalCmd = append(p.c.NormalCmd, cmd)
	} else {
		p.c.SuperCmd = append(p.c.SuperCmd, cmd)
	}
	return matcher
}

// 检查并添加必要的Rule
func (p *PluginProxy) checkRules(rules ...zero.Rule) []zero.Rule {
	// 是否为超级用户专属插件
	if p.c.IsSuperOnly {
		return append(rules, zero.SuperUserPermission)
	}
	return rules
}

// ---- 配置 ----

// AddConfig 添加配置
func (p *PluginProxy) AddConfig(key string, defaultValue interface{}) {
	p.u.addConfig(fmt.Sprintf("plugins.%s", p.key), key, defaultValue)
}

// GetConfig 获取配置
func (p *PluginProxy) GetConfig(key string) interface{} {
	return p.u.getConfig(fmt.Sprintf("plugins.%s", p.key), key)
}

// GetConfigString 获取String配置
func (p *PluginProxy) GetConfigString(key string) string {
	return cast.ToString(p.GetConfig(key))
}

// GetConfigInt64 获取Int64配置
func (p *PluginProxy) GetConfigInt64(key string) int64 {
	return cast.ToInt64(p.GetConfig(key))
}

// GetConfigFloat64 获取Float64配置
func (p *PluginProxy) GetConfigFloat64(key string) float64 {
	return cast.ToFloat64(p.GetConfig(key))
}

// GetConfigBool 获取Bool配置
func (p *PluginProxy) GetConfigBool(key string) bool {
	return cast.ToBool(p.GetConfig(key))
}

// ---- 数据库 ----

// GetDB 获取数据库
func (p *PluginProxy) GetDB() *gorm.DB {
	return p.u.db
}
