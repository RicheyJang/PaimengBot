package Myinfo

import (
	"encoding/json"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/HiOSU/UserInfo"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/wdvxdr1123/ZeroBot/message"

	//"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"io/ioutil"
	"net/http"
	"strings"
)

var info = manager.PluginInfo{
	Name: "OSU用户信息",
	Usage: `
	我的信息[0] : 查看自己的账号信息，后面的数字是查询的模式，
			   0 = osu!, 1 = Taiko, 
			   2 = CtB, 3 = osu!mania.
               默认值为0
	第一次使用请先 "OSU用户绑定" 来绑定OSUid
	`,
}

type User struct {
	UserID      string //"user_id"         //数字ID  0
	UserName    string //"username"       //名称 1
	JoinDate    string //"join_date"       //加入时间  2
	Country     string //"country"         //国家   18
	GlobalRank  string //"pp_rank"         //国际PP排名  9
	CountryRank string //"pp_country_rank" // 国内的PP排名  20
}

var proxy *manager.PluginProxy // 声明插件代理变量
func init() {
	proxy = manager.RegisterPlugin(info) //使用插件信息初始化插件代理
	if proxy == nil {                    // 若初始化失败，请return，失败原因会在日志中打印
		return //// 若初始化失败，请return，失败原因会在日志中打印
	}
	proxy.OnCommands([]string{"我的信息"}).SetBlock(true).Handle(MineInfo) //查看自己的账号信息
}

func MineInfo(ctx *zero.Ctx) {

	UserInfo.GetOsuid(ctx.Event.UserID) //查询数据表中用户绑定信息

	model := strings.TrimSpace(utils.GetArgs(ctx)) //获取用户要查询的模式
	Model := GetModel(model)                       //获取游戏模式

	//OSU获取用户资料API,其中 k 是 API的key , u 是 查询用户的纯数字ID 或者  ID ， model 是 查询的模式
	OsuAPI := "https://osu.ppy.sh/api/get_user" + "?k=51b88dd53687332618935b74d5a3bf22c8326826" + "&u=" + UserInfo.OSUid + "&m=" + model

	info := GetMyInfo(OsuAPI)

	if (len(info)) == 0 {
		ctx.Send("没有绑定OSU账号的说(○｀ 3′○)")
	} else {

		var USER User //创建一个新的对象
		USER.UserName = info[0]["username"]
		USER.UserID = info[0]["user_id"]
		USER.JoinDate = info[0]["join_date"]
		USER.Country = info[0]["country"]
		USER.CountryRank = info[0]["pp_country_rank"]
		USER.GlobalRank = info[0]["pp_rank"]

		if USER.CountryRank == "" {
			ctx.Send("从来没有玩过" + Model + "模式的说(～o￣3￣)～")
		} else {

			//去除时间后面的小时,分钟,秒
			//2020-08-17 23:02:42 ---->  2020-08-17
			comma := strings.Index(USER.JoinDate, " ")

			USER.JoinDate = USER.JoinDate[:comma]

			Image, _ := ToImageUser(USER, Model)

			ctx.Send(Image)

		}

	}

}

func GetMyInfo(API string) []map[string]string {

	result, err := http.Get(API)
	Json, _ := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil
	}

	Result := make([]map[string]string, 0) //创建一个Map

	_ = json.Unmarshal(Json, &Result) //将Json转化为Map

	return Result //返回一个map

}

func GetModel(ModelNumber string) string {

	var Model string
	if ModelNumber == "0" {

		Model = "osu!"
	} else if ModelNumber == "1" {
		Model = "Taiko"
	} else if ModelNumber == "2" {
		Model = "CtB"
	} else if ModelNumber == "3" {
		Model = "Osu!Mania"
	}
	return Model
}

func ToImageUser(user User, Model string) (message.MessageSegment, error) { ////生成图片(需要修改)

	windth := float64(400)
	height := float64(220)

	var dc = images.NewImageCtx(int(windth), int(height))

	dc.SetHexColor("#FFFFF0") // 设置画笔颜色为绿色
	dc.Clear()                // 使用当前颜色（绿）填满画布，即设置背景色

	if err := dc.LoadFontFace("./ttf/zh-cn.ttf", 30); err != nil { // 从本地加载字体文件
		panic(err)
	}
	dc.SetRGB(0, 0, 0) // 设置画笔颜色为黑色

	dc.DrawString("用户名: "+user.UserName, 2, 30*1)
	dc.DrawString("ID: "+user.UserID, 2, 30*2)
	dc.DrawString("模式: "+Model, 2, 30*3)
	dc.DrawString("注册时间: "+user.JoinDate, 2, 30*4)
	dc.DrawString("国内排名: "+user.CountryRank, 2, 30*5)
	dc.DrawString("国际排名: "+user.GlobalRank, 2, 30*6)
	dc.DrawString("国家/地区:"+user.Country, 2, 30*7)

	return dc.GenMessageAuto()
}

/*[
  	{
  		"user_id":"18141351",  //数字ID  0
  		"username":"DUNKEL233",   //名称 1
  		"join_date":"2020-08-17 23:02:42",  //加入时间  2

  		"count300":"167210",  3
  		"count100":"52889",  4
  		"count50":"1676",  5

  		"playcount":"392",  6
  		"ranked_score":"62309556",// 所有Ranked，Approved，Loved谱面中的最高分计数  7
  		"total_score":"130903735",  8

  		"pp_rank":"180525",//国际PP排名  9
  		"level":"27.2272",  10
  		"pp_raw":"822.3",  11
  		"accuracy":"93.36895751953125",  12

  		"count_rank_ss":"0",  // 获得的SS S A的歌曲数目  13
  		"count_rank_ssh":"0",  14
  		"count_rank_s":"8",  15
  		"count_rank_sh":"0",  16
  		"count_rank_a":"44",  17

  		"country":"CN",  //国家   18
  		"total_seconds_played":"30243",  19
  		"pp_country_rank":"8427", // 国内的PP排名  20
  		"events":[]
  	}
  ]
*/
