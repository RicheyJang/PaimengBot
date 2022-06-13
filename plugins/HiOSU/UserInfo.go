package HiOSU

/**
这个OSU查分功能是通过是实现OSU!API的v1版本实现的。
*/

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// var OSUid string //全局变量OSU id

func BindOSUidHandler(ctx *zero.Ctx) {
	OSUid := strings.TrimSpace(utils.GetArgs(ctx))
	if err := PutOsuID(ctx.Event.UserID, OSUid); err != nil { //有报错返回写入日志
		log.Errorf("PutUserUid err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("绑定成功ˋ( ° ▽、° ) ")
}

func ReferOSUidHandler(ctx *zero.Ctx) { //查询绑定的OSU Id
	ID := GetOsuid(ctx.Event.UserID) //查询数据表中用户绑定信息
	if ID == "" {
		ctx.Send("未绑定任何OSU账号")
	} else {
		ctx.Send("当前绑定的账号为:" + ID)
	}
}

func GetOsuid(id int64) (u string) {
	key := fmt.Sprintf("hiosu.UserId.U%v", id) //查询的key
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	return string(v)
}

func PutOsuID(id int64, u string) error { //将OSU id 和用户的ID写入数据表
	key := fmt.Sprintf("hiosu.UserId.U%v", id)                 //将表的键值定位 hiosu.UserId.U + Userid
	return proxy.GetLevelDB().Put([]byte(key), []byte(u), nil) //写入Key和aValue的值
}
