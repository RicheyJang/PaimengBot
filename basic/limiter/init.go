package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "插件CD限流",
	Usage: `防止频繁调用、刷屏；可以通过配置便携地设置CD时长，各用户CD互相独立
config-plugin配置项：
	只需配置config-plugin文件中的 插件Key.cd 配置项，就可以设置指定插件的CD时长了
	例如将 translate.cd 配置项值设为5s，则单个用户两次使用翻译插件的间隔将不允许低于5秒

	limiter.globalcd: 全局CD时长
	limiter.globalburst: 所有限流器的默认桶容量，请勿改动
	limiter.tip: CD中时是(true)否(false)提示调用者剩余CD时长`,
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
	proxy.AddConfig("tip", true)
	manager.AddPreHook(limiterHook)
	manager.WhenConfigFileChange(checkAllPluginsHook)
}

// 获取指定插件的PluginLimiter
func getPluginLimiter(plugin string) *PluginLimiter {
	// 维护插件Map
	v, ok := plMap.Load(plugin)
	if !ok {
		// 获取插件CD配置
		initCD := getCDConfig(plugin)
		if initCD == "" {
			return nil
		}
		// 创建PluginLimiter
		cd, err := time.ParseDuration(initCD)
		if err != nil || cd <= 0 { // CD解析失败
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

func limiterHook(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	// 全局限流检查
	if ctx.Event != nil && ctx.Event.PostType == "message" {
		globalLimiter := getPluginLimiter("limiter")
		if !couldUseAndTip(ctx, globalLimiter, ctx.Event.UserID) {
			log.Warnf("limiter：用户%v频率超出全局限流", ctx.Event.UserID)
			return fmt.Errorf("limiter：频率超出全局限流")
		}
	}
	// 插件限流检查
	pl := getPluginLimiter(condition.Key)
	if !couldUseAndTip(ctx, pl, ctx.Event.UserID) {
		log.Warnf("limiter：用户%v频率超出<%v>插件限流", ctx.Event.UserID, condition.Key)
		return fmt.Errorf("limiter：频率超出<%v>插件限流", condition.Key)
	}
	return nil
}

// 检查是否超限并提示
func couldUseAndTip(ctx *zero.Ctx, pl *PluginLimiter, id int64) bool {
	if pl == nil {
		return true
	}
	ok, left := pl.Allow(id)
	if !ok && left > time.Second && proxy.GetConfigBool("tip") { // 提示剩余时长
		if utils.IsMessagePrimary(ctx) {
			ctx.Send(message.Text(fmt.Sprintf("CD中，剩余%s", left.Round(time.Second))))
		} else {
			ctx.SendChain(message.At(id), message.Text(fmt.Sprintf("CD中，剩余%s", left)))
		}
	}
	return ok
}

// 获取指定插件的CD配置
func getCDConfig(plugin string) string {
	// 指定插件的CD配置项Key
	cdKey := consts.PluginConfigCDKey
	if plugin == "limiter" { // limiter插件本身的配置Key为全局CD
		cdKey = "globalCD"
	}
	plCDV := proxy.GetPluginConfig(plugin, cdKey)
	if plCDV == nil { // 没有设置CD
		return ""
	}
	return cast.ToString(plCDV)
}

// 检查插件的CD设置是否需要更新
func checkPluginCD(pl *PluginLimiter, newCD time.Duration) {
	if pl == nil || newCD <= 0 {
		return
	}
	if newCD != pl.GetCD() {
		pl.ResetCD(newCD)
		log.Infof("成功更新<%v>的CD：%s", pl.Key, newCD)
	}
	return
}

// Hook: 检查所有插件的限制器CD配置是否更新
func checkAllPluginsHook(event fsnotify.Event) error {
	plMap.Range(func(k, v interface{}) bool {
		key := k.(string)
		pl := v.(*PluginLimiter)
		if pl == nil {
			plMap.Delete(k)
			return true
		}
		// 获取当前CD配置
		cdStr := getCDConfig(key)
		if len(cdStr) == 0 { // CD被清空
			plMap.Delete(k)
			return true
		}
		// 解析CD
		cd, err := time.ParseDuration(cdStr)
		if err != nil {
			log.Warnf("%v插件CD设置出错，无法解析%v，err: %v", pl.Key, cdStr, err)
			return true
		}
		if cd <= 0 { // CD被设为0
			plMap.Delete(k)
			return true
		}
		// 检查CD是否更新
		checkPluginCD(pl, cd)
		return true
	})
	return nil
}
