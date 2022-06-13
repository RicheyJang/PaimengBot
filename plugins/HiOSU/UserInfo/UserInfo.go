package UserInfo

/**
这个OSU查分功能是通过是实现OSU!API的v1版本实现的。
*/

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"strings"
)

var info = manager.PluginInfo{
	Name: "OSU账号绑定",
	Usage: `
		账号绑定OSU [ID] : 
			绑定OSU账号,后面输入自己的纯数字ID或者你的OSU用户名,不填则取消绑定
		OSU账号查看 :
				查看该账号绑定的ID
	`,
}
var proxy *manager.PluginProxy // 声明插件代理变量

func init() {

	proxy = manager.RegisterPlugin(info) //使用插件信息初始化插件代理

	if proxy == nil { // 若初始化失败，请return，失败原因会在日志中打印
		return // 若初始化失败，请return，失败原因会在日志中打印
	}

	proxy.OnCommands([]string{"账号绑定OSU", "账号绑定osu", "账号绑定Osu", "账号绑定OSu"}).SetBlock(true).Handle(GetOSUid) //绑定账号
	proxy.OnCommands([]string{"OSU账号查看", "账号查看"}).SetBlock(true).Handle(ReferOSUid)                        //账号查看

}

var OSUid string //全局变量OSU id

func GetOSUid(ctx *zero.Ctx) {

	OSUid = strings.TrimSpace(utils.GetArgs(ctx))

	if err := PutOsuID(ctx.Event.UserID, OSUid); err != nil { //有报错返回写入日志
		ctx.Send("失败了...")
		log.Errorf("PutUserUid err: %v", err)
		return
	}

	ctx.Send("绑定成功ˋ( ° ▽、° ) ")
}
func GetOsuid(id int64) (u string) {

	key := fmt.Sprintf("UserId.U%v", id) //查询的key

	v, err := proxy.GetLevelDB().Get([]byte(key), nil)

	if err != nil {
		return
	}

	_ = json.Unmarshal(v, &OSUid)

	return OSUid

}

func PutOsuID(id int64, u string) error { //将OSU id 和用户的ID写入数据表

	key := fmt.Sprintf("UserId.U%v", id) //将表的键值定位 genshin_uid.u + Userid

	value, err := json.Marshal(u) //???

	if err != nil {
		return err
	}

	return proxy.GetLevelDB().Put([]byte(key), value, nil) //写入Key和aValue的值

}

func ReferOSUid(ctx *zero.Ctx) { //查询绑定的OSU Id

	ID := GetOsuid(ctx.Event.UserID) //查询数据表中用户绑定信息

	if ID == "" {

		ctx.Send("未绑定任何OSU账号")

	} else {

		ctx.Send("当前绑定的账号为:" + ID)

	}

}
