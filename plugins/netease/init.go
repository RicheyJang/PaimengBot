package netease

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "网易云评论",
	Usage: `用法：
	网易云评论/锐评 随机给出一条网易云评论`,
	Classify: "一般功能",
}

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"网易云评论", "锐评"}).SetBlock(true).SecondPriority().Handle(getComment)

}

func getComment(ctx *zero.Ctx) {
	var c = client.NewHttpClient(nil)
	json, err := c.GetGJson("https://api.vvhan.com/api/reping")
	if err != nil || !json.Get("success").Bool() {
		log.Warnf("reping err: user=%v,url=%v,err=%v", ctx.Event.UserID, err)
	}
	ctx.Send(message.Text(json.Get("data").Get("content")))

}
