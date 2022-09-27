package withdraw

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/rules"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name:     "撤回",
	Classify: "群功能",
	Usage: `用法：
	回复需要撤回的消息"撤回"，即可让Bot撤回该消息`,
	AdminLevel: 5,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(zero.OnlyGroup, rules.ReplyAndCommands("撤回", "快撤回")).SetBlock(true).SecondPriority().Handle(withDrawMsg)
}

func withDrawMsg(ctx *zero.Ctx) {
	replyID := cast.ToInt64(ctx.State["reply_id"])
	if replyID == 0 {
		log.Warn("get reply id = 0")
		return
	}
	msg := ctx.GetMessage(message.NewMessageIDFromInteger(cast.ToInt64(replyID)))
	// 检查是否为机器人本人消息
	//if msg.Sender != nil && msg.Sender.ID != 0 && msg.Sender.ID != ctx.Event.SelfID {
	//	ctx.Send(fmt.Sprintf("%v只能撤回%v自己发出的消息哦", utils.GetBotNickname(), utils.GetBotNickname()))
	//	return
	//}
	ctx.DeleteMessage(message.NewMessageIDFromInteger(cast.ToInt64(replyID)))
	log.Infof("撤回消息 %v (id=%v)", utils.JsonString(msg), replyID)
}
