package genshin_query

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_cookie"
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
	// 获取用户cookie
	user_cookie := genshin_cookie.GetUserCookie(ctx.Event.UserID)
	user_uid := genshin_cookie.GetUserUid(ctx.Event.UserID)
	cookie_msg := `用户未设置正确的cookie或uid:
	设置方法：下载应急食品
	通过应急食品获取cookie后 假设cookie是 xxx!sdasdxsadx 
	则私聊机器人:
		存入cookie xxx!sdasdxsadx
		存入uid 110722321
	请注意 存入cookie和存入uid后面没有冒号 是空格 然后是你的内容
	cookie详细获取方法：打开应急食品 进入工具 进入管理米游社账号 添加账号 （登录你的账号） 长按你登录成功的账号即可复制
	存入cookie和存入uid是俩条不同的命令，需要分开发`
	if len(user_cookie) <= 10 {
		ctx.Send(images.GenStringMsg(cookie_msg + "\ncookie设置失败"))
		return
	}
	if len(user_uid) <= 5 {
		ctx.Send(images.GenStringMsg(cookie_msg + "\nuid设置失败"))
		// ctx.Send(images.GenStringMsg("用户cookie为空 请存入cookie\n存入方法:\n私聊机器人\"存入cookie 自己的cookie\"\n\"存入uid 自己的uid\""))
		return
	}
	msg, _, _ := Query(user_uid, user_cookie)
	fmt.Printf("查询状态%s", msg)
	ctx.Send(message.Text(fmt.Sprintf("查询:%s", msg)))
}
