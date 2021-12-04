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
	p.addCommands([]string{reg}, rules...)
	return matcher
}

// OnCommands 添加新的命令匹配器
func (p *PluginProxy) OnCommands(cmd []string, rules ...zero.Rule) *zero.Matcher {
	rules = p.checkRules(rules...)
	matcher := p.u.engine.OnCommandGroup(cmd, rules...)
	p.addCommands(cmd, rules...)
	return matcher
}

func (p *PluginProxy) addCommands(cmd []string, rules ...zero.Rule) {
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
	p.c.InitialCron()
	id, err = p.c.schedule.AddFunc(spec, fn)
	if err == nil { // 尝试开启定时任务
		log.Infof("<%v>成功添加定时任务 spec: %v,id: %v", p.key, spec, id)
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

// AddScheduleOnceFunc 便携添加定时任务，在等待period（period<1年）时长后执行仅一次
func (p *PluginProxy) AddScheduleOnceFunc(period time.Duration, fn func()) (id cron.EntryID, err error) {
	// 立即执行
	if period <= 0 {
		go fn()
		return 0, nil
	}
	// 一分钟以内
	if period <= time.Minute {
		go func() {
			time.Sleep(period)
			fn()
		}()
		log.Infof("<%v>成功添加定时任务 after: %v,no id", p.key, period)
		return 0, nil
	}
	// 大于1年
	if period >= time.Hour*24*365 {
		return 0, fmt.Errorf("too long duration: %v", period)
	}
	// 一分钟以上，使用Cron
	p.c.InitialCron()
	id = p.c.schedule.Schedule(cron.Every(period), cron.FuncJob(func() {
		p.c.schedule.Remove(id)
		log.Debugf("Auto Remove <%v> Job : %v", p.key, id)
		fn()
	}))
	log.Infof("<%v>成功添加定时任务 after: %v,id: %v", p.key, period, id)
	return id, err
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
	p.u.addConfig(fmt.Sprintf("%s", p.key), key, defaultValue)
}

// GetConfig 获取配置
func (p *PluginProxy) GetConfig(key string) interface{} {
	return p.u.getConfig(fmt.Sprintf("%s", p.key), key)
}

// GetPluginConfig 获取其它插件的指定配置
func (p *PluginProxy) GetPluginConfig(plugin string, key string) interface{} {
	return p.u.getConfig(fmt.Sprintf("%s", plugin), key)
}

// AddAPIConfig 添加API配置（仅限String类型）
func (p *PluginProxy) AddAPIConfig(key string, defaultValue string) {
	p.u.addConfig("api", key, defaultValue)
}

// GetAPIConfig 获取API配置
func (p *PluginProxy) GetAPIConfig(key string) string {
	return cast.ToString(p.GetPluginConfig("api", key))
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

// GetConfigStrings 获取[]string配置
func (p *PluginProxy) GetConfigStrings(key string) []string {
	return cast.ToStringSlice(p.GetConfig(key))
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
