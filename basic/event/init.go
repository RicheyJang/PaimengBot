package event

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:     "处理除消息外其它基本事件",
	Usage:    "防止被动拉入群聊；捕获好友、群邀请发送给超级用户",
	IsHidden: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRequest().SetBlock(true).FirstPriority().Handle(handleInvite) // 捕获好友、群邀请发送给超级用户
	proxy.OnNotice(utils.CheckDetailType("group_increase"), func(ctx *zero.Ctx) bool {
		return ctx.Event.SelfID == ctx.Event.UserID
	}).SetBlock(true).FirstPriority().Handle(preventForcedInviteGroup) // 防止被动拉入群聊
}

// 机器人初入群聊时
func preventForcedInviteGroup(ctx *zero.Ctx) {

}

// 收到邀请入群、加好友请求时
func handleInvite(ctx *zero.Ctx) {

}
