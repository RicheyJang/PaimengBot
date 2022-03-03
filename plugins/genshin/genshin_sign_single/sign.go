package genshin_sign_single

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_public"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "签到",
	Usage: `如果你填写了对应的cookie
将会自动在查询对应的信息 说 签到 就可以啦
` + genshin_public.GetInitializaationPrompt(),
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info) // [3] 使用插件信息初始化插件代理
	if proxy == nil {                    // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	// [4] 此处进行其它初始化操作
	proxy.OnCommands([]string{"签到"}).SetBlock(true).SetPriority(3).Handle(sign)
}

// [5] 其它代码实现

func sign(ctx *zero.Ctx) {
	user_cookie, user_uid, cookie_msg, err := genshin_public.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookie_msg))
		return
	}
	msg, err := genshin_sign.Sign(user_uid, user_cookie)
	fmt.Printf("签到%s", msg)
	ctx.Send(message.Text(fmt.Sprintf("签到:%s", msg)))
}
