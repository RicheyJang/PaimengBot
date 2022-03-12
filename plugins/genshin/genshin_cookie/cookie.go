package genshin_cookie

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var Info = manager.PluginInfo{
	Name: "米游社管理",
	Usage: `如何绑定米游社cookie和原神uid：
	原神绑定cookie [你的cookie]：cookie是重要信息，请务必在私聊中使用
	原神绑定uid [你的uid]
如何解绑：
	使用上述命令，不填参数([你的cookie]和[你的uid])即可`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(Info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"原神绑定cookie", "米游社绑定cookie"}).SetBlock(true).SetPriority(3).Handle(saveCookie)
	proxy.OnCommands([]string{"原神绑定uid"}).SetBlock(true).SetPriority(3).Handle(saveUid)
}

func saveCookie(ctx *zero.Ctx) {
	if !utils.IsMessagePrimary(ctx) {
		ctx.Send("请在私聊中存放cookie！快撤回！")
		return
	}
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if err := PutUserCookie(ctx.Event.UserID, args); err != nil {
		ctx.Send("失败了...")
		log.Errorf("PutUserCookie err: %v", err)
		return
	}
	ctx.Send("好哒")
}

func saveUid(ctx *zero.Ctx) {
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if err := PutUserUid(ctx.Event.UserID, args); err != nil {
		ctx.Send("失败了...")
		log.Errorf("PutUserUid err: %v", err)
		return
	}
	ctx.Send("好哒")
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
