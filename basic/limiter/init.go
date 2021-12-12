package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:        "插件CD限流",
	Usage:       "防止频繁调用、刷屏；可以通过配置便携地设置CD",
	IsPassive:   true,
	IsSuperOnly: true,
}
var plMap sync.Map

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.AddConfig("globalCD", "350ms")
	proxy.AddConfig("globalBurst", 1)
	manager.AddPreHook(limiterHook)
}

func getPluginLimiter(plugin, initCD string) *PluginLimiter {
	// 维护插件Map
	v, ok := plMap.Load(plugin)
	if !ok {
		// 创建PluginLimiter
		cd, err := time.ParseDuration(initCD)
		if err != nil { // CD解析失败
			return nil
		}
		burst := proxy.GetConfigInt64("globalBurst")
		if burst <= 0 {
			burst = 1
		}
		pl := NewPluginLimiter(cd, int(burst))
		pl.Key = plugin
		log.Infof("创建<%v>的PluginLimiter, CD=%v", plugin, cd)
		// 存储
		plMap.Store(plugin, pl)
		return pl
	}
	return v.(*PluginLimiter)
}

// 检查插件的CD设置是否需要更新
func checkPluginCD(pl *PluginLimiter, newCD string) {
	if pl == nil || len(newCD) == 0 {
		return
	}
	cd, err := time.ParseDuration(newCD)
	if err != nil {
		log.Warnf("CD设置出错，无法解析%v，err: %v", newCD, err)
	}
	if cd != pl.GetCD() {
		pl.ResetCD(cd)
		log.Infof("成功更新<%v>的CD：%v", pl.Key, newCD)
	}
}

func limiterHook(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	// 全局限流检查
	if ctx.Event != nil && ctx.Event.PostType == "message" {
		globalLimiter := getPluginLimiter("limiter", proxy.GetConfigString("globalCD"))
		checkPluginCD(globalLimiter, proxy.GetConfigString("globalCD"))
		if globalLimiter != nil && !globalLimiter.Allow(ctx.Event.UserID) {
			log.Warnf("limiter：用户%v频率超出全局限流", ctx.Event.UserID)
			return fmt.Errorf("limiter：频率超出全局限流")
		}
	}
	// 插件限流检查
	plCDV := proxy.GetPluginConfig(condition.Key, consts.PluginConfigCDKey)
	if plCDV == nil { // 未设置CD
		return nil
	}
	plCD := cast.ToString(plCDV)
	if len(plCD) == 0 { // CD被清空
		return nil
	}
	// 设置了CD
	pl := getPluginLimiter(condition.Key, plCD)
	checkPluginCD(pl, plCD)
	if pl != nil && !pl.Allow(ctx.Event.UserID) {
		log.Warnf("limiter：用户%v频率超出<%v>插件限流", ctx.Event.UserID, condition.Key)
		return fmt.Errorf("limiter：频率超出<%v>插件限流", condition.Key)
	}
	return nil
}
