package genshin_query

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_public"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "查询信息",
	Usage: `如果你填写了对应的cookie
将会自动在查询对应的信息 说 查询 就可以啦
存入方法:
私聊机器人"存入cookie 自己的cookie"
"存入uid 自己的uid"`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info) // [3] 使用插件信息初始化插件代理
	if proxy == nil {                    // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	// [4] 此处进行其它初始化操作
	proxy.OnCommands([]string{"查询", "体力", "树脂"}).SetBlock(true).SetPriority(3).Handle(queryInfo)
}

// [5] 其它代码实现

func queryInfo(ctx *zero.Ctx) {
	user_uid, user_cookie, cookie_msg, err := genshin_public.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookie_msg))
		return
	}
	msg, _, _ := Query(user_uid, user_cookie)
	ctx.Send(message.Text(fmt.Sprintf("查询:%s", msg)))
}
