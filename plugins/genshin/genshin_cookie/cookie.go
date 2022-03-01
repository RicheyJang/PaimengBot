package genshin_cookie

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "cookie管理",
	Usage: `存入cookie
存入uid`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info) // [3] 使用插件信息初始化插件代理
	if proxy == nil { // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	// [4] 此处进行其它初始化操作
	proxy.OnCommands([]string{"存入cookie"}).SetBlock(true).SetPriority(3).Handle(saveCookie)
	proxy.OnCommands([]string{"存入uid"}).SetBlock(true).SetPriority(3).Handle(saveUid)
}

// [5] 其它代码实现

func saveCookie(ctx *zero.Ctx) {
	args := utils.GetArgs(ctx)
	PutUserCookie(ctx.Event.UserID,args)
	// 获取用户cookie
	user_cookie := GetUserCookie(ctx.Event.UserID)
	if user_cookie == "" {
		ctx.Send(message.Text(fmt.Sprintf("用户cookie为空 请存入cookie")))
		return
	}
	ctx.Send(message.Text(fmt.Sprintf("用户cookie为:%s",user_cookie)))
}
func saveUid(ctx *zero.Ctx) {
	args := utils.GetArgs(ctx)
	PutUserUid(ctx.Event.UserID,args)
	// 获取用户cookie
	uid := GetUserUid(ctx.Event.UserID)
	if uid == "" {
		ctx.Send(message.Text(fmt.Sprintf("用户uid为空 请存入uid")))
		return
	}
	ctx.Send(message.Text(fmt.Sprintf("用户uid为:%s",uid)))
}


func GetUserCookie(id int64) (u string) {
	key := fmt.Sprintf("genshin_cookie.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	_ = json.Unmarshal(v, &u)
	return
}
func GetUserUid(id int64) (u string) {
	key := fmt.Sprintf("genshin_uid.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	_ = json.Unmarshal(v, &u)
	return
}

func PutUserCookie(id int64, u string) error {
	key := fmt.Sprintf("genshin_cookie.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}
func PutUserUid(id int64, u string) error {
	key := fmt.Sprintf("genshin_uid.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}