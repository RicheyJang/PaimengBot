package genshin_sign

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

//// PluginInfo 插件信息
//type PluginInfo struct {
//	Name        string // [必填] 插件名称
//	Usage       string // [必填] 插件用法描述：会在插件帮助详情中展示
//
//	SuperUsage  string // [选填] 插件超级用户用法描述
//	Classify    string // [选填] 插件分类，为空时代表默认分类
//	IsPassive   bool   // [选填] 是否为被动插件：在帮助中被标识为被动功能；
//	IsSuperOnly bool   // [选填] 是否为超级用户专属插件：若true，消息性事件会自动加上SuperOnly检查；在帮助中只有超级用户私聊可见；
//	AdminLevel  int    // [选填] 群管理员使用最低级别： 0 表示非群管理员专用插件 >0 表示数字越低，权限要求越高；会在帮助中进行标识；配置文件中 插件Key.adminlevel 配置项优先级高于此项
//}

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
	if proxy == nil { // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	// [4] 此处进行其它初始化操作
	proxy.OnCommands([]string{"查询", "体力","树脂"}).SetBlock(true).SetPriority(3).Handle(queryInfo)
}

// [5] 其它代码实现

func queryInfo(ctx *zero.Ctx) {
	// 获取用户cookie
	user_cookie := GetUserCookie(ctx.Event.UserID)
	if user_cookie == "" {
		ctx.Send(message.Text(fmt.Sprintf("用户cookie为空 请存入cookie\n存入方法:\n私聊机器人\"存入cookie 自己的cookie\"\n\"存入uid 自己的uid\"")))
		return
	}
	ctx.Send(message.Text(fmt.Sprintf("查询:%s", query_func(GetUserUid(ctx.Event.UserID),GetUserCookie(ctx.Event.UserID)))));
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