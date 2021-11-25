package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"
)

// PluginProxy 插件代理，呈现给插件，用于添加事件动作、读写配置、获取插件锁、添加定时任务
// 插件在注册后，应只与此代理交互，与Manager再无交际
type PluginProxy struct {
	key      string         // 插件Key
	u        *PluginManager // 所从属的插件管理器
	userLock sync.Map       // 用户锁

	c PluginCondition // 插件状态（被管理器控制）
}

// ---- 事件动作 ----

// On 添加新的指定消息类型的匹配器
// zero.Engine.On的附加处理拷贝
func (p *PluginProxy) On(tp string, rules ...zero.Rule) *zero.Matcher {
	if tp == "message" {
		rules = p.checkRules(rules...)
	}
	matcher := p.u.engine.On(tp, rules...)
	return matcher
}

// OnRegex 添加新的正则匹配器
func (p *PluginProxy) OnRegex(reg string, rules ...zero.Rule) *zero.Matcher {
	rules = p.checkRules(rules...)
	matcher := p.u.engine.OnRegex(reg, rules...)
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

func (p *PluginProxy) OnRequest(rules ...zero.Rule) *zero.Matcher {
	return p.On("request", rules...)
}

func (p *PluginProxy) OnNotice(rules ...zero.Rule) *zero.Matcher {
	return p.On("notice", rules...)
}

// ---- 定时任务 ----

// AddScheduleFunc 添加定时任务，并自动启动
func (p *PluginProxy) AddScheduleFunc(spec string, fn func()) (id cron.EntryID, err error) {
	if p.c.schedule == nil {
		p.c.schedule = cron.New(cron.WithLogger(utils.NewCronLogger()), // 设置日志
			cron.WithChain(cron.SkipIfStillRunning(utils.NewCronLogger()))) // 若前一任务仍在执行，则跳过当前任务
	}
	id, err = p.c.schedule.AddFunc(spec, fn)
	if err == nil { // 尝试开启定时任务
		p.c.StartCron()
		log.Infof("%v成功添加定时任务 spec: %v", p.key, spec)
	} else {
		log.Errorf("%v添加定时任务失败, err: %v", p.key, err)
	}
	return
}

// AddScheduleEveryFunc 便携添加定时任务，固定时间间隔执行，duration符合time.ParseDuration
func (p *PluginProxy) AddScheduleEveryFunc(duration string, fn func()) (id cron.EntryID, err error) {
	spec := fmt.Sprintf("@every %s", duration)
	return p.AddScheduleFunc(spec, fn)
}

// AddScheduleDailyFunc 便携添加定时任务，每天hour:minute时执行
func (p *PluginProxy) AddScheduleDailyFunc(hour, minute int, fn func()) (id cron.EntryID, err error) {
	spec := fmt.Sprintf("%d %d * * *", minute, hour)
	return p.AddScheduleFunc(spec, fn)
}

// DeleteSchedule 删除定时任务
func (p *PluginProxy) DeleteSchedule(id cron.EntryID) {
	if p.c.schedule == nil {
		return
	}
	p.c.schedule.Remove(id)
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
	return p.u.GetDB()
}

// ---- 插件锁 ----

// LockUser 对用户上锁，返回能否继续操作，true表示该用户正在被锁定，false表示该用户未被锁定、可以进行下一步操作，同时会对其上锁
func (p *PluginProxy) LockUser(userID int64) bool {
	_, ok := p.userLock.Load(userID)
	// 上锁 or 更新获得锁时间
	now := time.Now().Unix()
	p.userLock.Store(userID, now)
	return ok
}

// UnlockUser 解锁用户
func (p *PluginProxy) UnlockUser(userID int64) {
	p.userLock.Delete(userID)
}
