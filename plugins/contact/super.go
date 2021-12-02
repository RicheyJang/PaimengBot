package contact

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "联系管理员",
	Usage: `用法：
	联系管理员 [XXX]：将消息XXX发给管理本机器人的管理者(超级用户)`,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"联系管理员", "联系超级用户", "嘀嘀嘀", "滴滴滴"}, zero.OnlyToMe).
		SetBlock(true).SecondPriority().Handle(contactSuper)
}

func contactSuper(ctx *zero.Ctx) {
	org := ctx.Event.Message
	var str string
	if utils.IsMessagePrimary(ctx) { // 私聊消息
		if ctx.Event.Sender != nil { // 有发送者
			str = fmt.Sprintf("收到来自用户%v(%v)的消息：\n", ctx.Event.Sender.NickName, ctx.Event.UserID)
		} else { // 无发送者
			log.Infof("消息Event：%v", utils.JsonString(ctx.Event))
			str = fmt.Sprintf("收到来自未知用户(%v)的消息：\n", ctx.Event.UserID)
		}
		ctx.Send("好哒")
	} else { // 群聊消息
		group := ctx.GetGroupInfo(ctx.Event.GroupID, false)
		if ctx.Event.SubType == "anonymous" && ctx.Event.Anonymous != nil { // 匿名
			ano := gjson.Parse(utils.JsonString(ctx.Event.Anonymous))
			str = fmt.Sprintf("收到来自群%v(%v)\n匿名用户%v(%v)的消息：\n",
				group.Name, ctx.Event.GroupID, ano.Get("name").String(), ano.Get("id").Int())
		} else if ctx.Event.Sender != nil { // 正常
			str = fmt.Sprintf("收到来自群%v(%v)\n用户%v(%v:%v)的消息：\n",
				group.Name, ctx.Event.GroupID, ctx.Event.Sender.NickName, ctx.Event.Sender.Card, ctx.Event.UserID)
		} else { // 未知
			str = fmt.Sprintf("收到来自群%v(%v)\n未知用户(%v)的消息：\n",
				group.Name, ctx.Event.GroupID, ctx.Event.UserID)
		}
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("好哒"))
	}
	str += "------------\n"
	utils.SendToSuper(append(message.Message{message.Text(str)}, org...)...)
}
