package keyword

import (
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/spf13/cast"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name:     "关键词撤回",
	Classify: "群功能",
	Usage: `若有群成员说了包含某些关键词的话，Bot将自动撤回该消息
需要Bot作为群管理员；配置关键词请联系超级用户`,
	SuperUsage: `config-plugin配置项：
	keyword.withdraw: 触发自动撤回的关键词列表，为字符串数组格式，默认为空
	keyword.skipsuper: 是(true)否(false)给予超级用户特权，不检查超级用户发出的消息`,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(zero.OnlyGroup, needCheck, needDeal).SetBlock(false).FirstPriority().Handle(msgHandler)
	proxy.AddConfig("withdraw", []string{}) // 触发撤回的关键词列表
	proxy.AddConfig("skipsuper", true)      // 是否给予超级用户特权
	// 配置文件更新时重载关键词列表
	manager.WhenConfigFileChange(func(event fsnotify.Event) error {
		flushWithdrawKeywords()
		return nil
	})
}

// 检查：是否需要检查关键词
func needCheck(ctx *zero.Ctx) bool {
	// 基本检查
	if !(len(getWithdrawKeywords()) > 0 && ctx.Event.MessageID != 0) {
		return false
	}
	// 消息来源检查
	if ctx.Event.SelfID == ctx.Event.UserID || (proxy.GetConfigBool("skipsuper") && utils.IsSuperUser(ctx.Event.UserID)) {
		return false
	}
	return true
}

// 检查：是否包含关键词
func needDeal(ctx *zero.Ctx) bool {
	str := ctx.MessageString()
	// 检查是否需要撤回
	withdrawList := getWithdrawKeywords()
	for _, v := range withdrawList {
		if len(v) == 0 {
			continue
		}
		if strings.Contains(str, v) { // 包含关键词
			ctx.State["keyword"] = v
			ctx.State["action"] = "withdraw"
			return true
		}
	}
	// 其它检查
	return false
}

// 对满足上述要求的消息进行处理
func msgHandler(ctx *zero.Ctx) {
	if ctx.State["action"] == "withdraw" {
		log.Infof(`触发关键词"%v"，撤回消息ID：%v`, ctx.State["keyword"], ctx.Event.MessageID)
		ctx.Block()
		ctx.Send("嘘！")
		ctx.DeleteMessage(message.NewMessageIDFromInteger(cast.ToInt64(ctx.Event.MessageID)))
	}
	// 其它处理
}

// 为提高性能，维护触发撤回的关键词列表
var withdrawKeywords []string

func flushWithdrawKeywords() {
	withdrawKeywords = proxy.GetConfigStrings("withdraw")
	if len(withdrawKeywords) > 0 {
		log.Infof("<keyword> 已载入%d个撤回触发关键词", len(withdrawKeywords))
	}
}

func getWithdrawKeywords() []string {
	return withdrawKeywords
}
