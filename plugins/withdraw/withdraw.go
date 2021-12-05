package withdraw

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/RicheyJang/PaimengBot/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "撤回",
	Usage: `用法：
	回复需要撤回的消息：撤回
	即可让Bot撤回该消息`,
	AdminLevel: 5,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(zero.OnlyGroup, zero.OnlyToMe, checkNeedWithdraw).SetBlock(true).FirstPriority().Handle(withDrawMsg)
}

func checkNeedWithdraw(ctx *zero.Ctx) bool {
	if len(ctx.Event.Message) < 2 || ctx.Event.Message[0].Type != "reply" {
		return false
	}
	for _, msg := range ctx.Event.Message {
		if msg.Type == "text" && checkTextMsgNeedWithdraw(msg.Data["text"]) {
			ctx.State["reply_id"] = ctx.Event.Message[0].Data["id"]
			return true
		}
	}
	return false
}

func checkTextMsgNeedWithdraw(msg string) bool {
	if len(msg) <= 1 {
		return false
	}
	msg = strings.TrimSpace(msg)
	for _, nick := range utils.GetBotConfig().NickName {
		if strings.HasPrefix(msg, nick) {
			msg = msg[len(nick):]
			msg = strings.TrimSpace(msg)
			break
		}
	}
	return strings.HasPrefix(msg, "撤回") || strings.HasPrefix(msg, "快撤回")
}

func withDrawMsg(ctx *zero.Ctx) {
	replyID := cast.ToInt64(ctx.State["reply_id"])
	if replyID == 0 {
		log.Warn("get reply id = 0")
		return
	}
	msg := ctx.GetMessage(replyID)
	if msg.Sender != nil && msg.Sender.ID != 0 && msg.Sender.ID != ctx.Event.SelfID {
		ctx.Send(fmt.Sprintf("%v只能撤回%v自己发出的消息哦", utils.GetBotNickname(), utils.GetBotNickname()))
		return
	}
	ctx.DeleteMessage(replyID)
	log.Infof("撤回消息 %v (id=%v)", utils.JsonString(msg), replyID)
}
