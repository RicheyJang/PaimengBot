package HiOSU

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/images"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type User struct {
	UserID      string `json:"user_id"`         //数字ID  0
	UserName    string `json:"username"`        //名称 1
	JoinDate    string `json:"join_date"`       //加入时间  2
	Country     string `json:"country"`         //国家   18
	GlobalRank  string `json:"pp_rank"`         //国际PP排名  9
	CountryRank string `json:"pp_country_rank"` // 国内的PP排名  20
}

func MineInfoHandler(ctx *zero.Ctx) {
	key := proxy.GetConfigString("key")
	if len(key) == 0 {
		ctx.Send("管理员尚未配置API KEY，快去催他！")
		return
	}
	//查询数据表中用户绑定信息
	OSUid := GetOsuid(ctx.Event.UserID)
	if len(OSUid) == 0 {
		ctx.Send("没有绑定OSU账号的说(○｀ 3′○)")
		return
	}
	//获取用户要查询的模式
	model := strings.TrimSpace(utils.GetArgs(ctx))
	if model != "1" && model != "2" && model != "3" {
		model = "0"
	}
	Model := GetModel(model)

	//OSU获取用户资料API,其中 k 是 API的key , u 是 查询用户的纯数字ID 或者  ID ， model 是 查询的模式
	OsuAPI := "https://osu.ppy.sh/api/get_user?k=" + key + "&u=" + OSUid + "&m=" + model
	USER, err := GetMyInfo(OsuAPI)
	if err != nil {
		log.Errorf("GetMyInfo err: %v", err)
		ctx.Send("失败了...")
		return
	}

	if USER.CountryRank == "" {
		ctx.Send("从来没有玩过" + Model + "模式的说(～o￣3￣)～")
	} else {
		//去除时间后面的小时,分钟,秒
		//2020-08-17 23:02:42 --->  2020-08-17
		comma := strings.Index(USER.JoinDate, " ")
		USER.JoinDate = USER.JoinDate[:comma]
		ctx.Send(ToImageUser(USER, Model))
	}
}

func GetMyInfo(API string) (User, error) {
	// 调用
	c := client.NewHttpClient(nil)
	r, err := c.GetReader(API)
	if err != nil {
		return User{}, err
	}
	defer r.Close()
	// 解析
	d := json.NewDecoder(r)
	var users []User
	if err = d.Decode(&users); err != nil {
		return User{}, err
	}
	if len(users) == 0 {
		return User{}, fmt.Errorf("user info is empty")
	}
	return users[0], nil
}

func GetModel(ModelNumber string) string {
	switch ModelNumber {
	case "1":
		return "Taiko"
	case "2":
		return "CtB"
	case "3":
		return "Osu!Mania"
	default:
		return "osu!"
	}
}

func ToImageUser(user User, Model string) message.MessageSegment { //生成图片(需要修改)
	str := "用户名: " + user.UserName +
		"\nID: " + user.UserID +
		"\n模式: " + Model +
		"\n注册时间: " + user.JoinDate +
		"\n国内排名: " + user.CountryRank +
		"\n国际排名: " + user.GlobalRank +
		"\n国家/地区:" + user.Country
	return images.GenStringMsg(str)
}
