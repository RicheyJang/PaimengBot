package RecentPlay

import (
	"encoding/json"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/HiOSU/Myinfo"
	"github.com/RicheyJang/PaimengBot/plugins/HiOSU/UserInfo"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var info = manager.PluginInfo{
	Name: "OSU最近成绩",
	Usage: `最近成绩 [0] :获取最新一次的成绩，后面的数字是查询的模式"
			   0 = osu!, 1 = Taiko, 
			   2 = CtB, 3 = osu!mania.
               默认值为0
	第一次使用请先 "OSU用户绑定" 来绑定OSUid
	`,
}

//   /api/get_user_recent

var proxy *manager.PluginProxy // 声明插件代理变量

func init() {

	proxy = manager.RegisterPlugin(info) //使用插件信息初始化插件代理

	if proxy == nil { // 若初始化失败，请return，失败原因会在日志中打印
		return //// 若初始化失败，请return，失败原因会在日志中打印
	}

	proxy.OnCommands([]string{"最近成绩"}).SetBlock(true).Handle(RecentPlay) //查看自己的账号信息

}

type Recent struct {
	BeatmapId string //谱面ID
	Score     string //分数
	MaxCombo  string //最大连击
	UserId    string //玩家ID
	Rank      string //评价
	Count300  string //GREAT数
	Count100  string //GOOD数
	Count50   string //BAD数
	CountMiss string //MISS数
	Date      string //游玩日期

}

func RecentPlay(ctx *zero.Ctx) {

	UserInfo.GetOsuid(ctx.Event.UserID) //查询数据表中用户绑定信息

	model := strings.TrimSpace(utils.GetArgs(ctx)) //获取用户要查询的模式
	Model := Myinfo.GetModel(model)                //获取游玩模式

	//获取最近游玩成绩API + 其中 k 是 API的key , u 是 查询用户的纯数字ID 或者  ID ， model 是 查询的模式 , limit 是返回的成绩个数
	OsuAPI := "https://osu.ppy.sh/api/get_user_recent" + "?k=51b88dd53687332618935b74d5a3bf22c8326826" + "&u=" + UserInfo.OSUid + "&m=" + model + "&limit=1"

	Recentplay := GetRecentPlay(OsuAPI)

	if len(Recentplay) == 0 {
		ctx.Send("最近没有玩过" + Model + "这个模式哦")
	} else {

		var recent Recent
		recent.UserId = Recentplay[0]["user_id"]
		recent.BeatmapId = Recentplay[0]["beatmap_id"]
		recent.Score = Recentplay[0]["score"]
		recent.Rank = Recentplay[0]["rank"]
		//recent.MaxCombo = Recentplay[0]["maxcombo"]
		//recent.Count300 = Recentplay[0]["count300"]
		//recent.Count100 = Recentplay[0]["count100"]
		//recent.Count50 = Recentplay[0]["count50"]
		//recent.CountMiss = Recentplay[0]["countmiss"]
		recent.Date = Recentplay[0]["date"]

		//comma := strings.Index(recent.Date, " ")

		//recent.Date = recent.Date[:comma]

		Image, _ := ToImageRecent(recent, Model, UserInfo.OSUid)

		ctx.Send(Image)

	}

}

func GetRecentPlay(API string) []map[string]string {

	result, err := http.Get(API)
	Json, _ := ioutil.ReadAll(result.Body)

	if err != nil {
		return nil
	}

	Result := make([]map[string]string, 0) //创建一个Map

	_ = json.Unmarshal(Json, &Result) //将Json转化为Map

	return Result //返回一个map
}

func ToImageRecent(recent Recent, Model string, OsuId string) (message.MessageSegment, error) { //生成图片(需要修改)

	windth := float64(400)
	height := float64(280)

	var dc = images.NewImageCtx(int(windth), int(height))

	dc.SetHexColor("#FFFFF0") // 设置画笔颜色为绿色
	dc.Clear()                // 使用当前颜色（绿）填满画布，即设置背景色

	if err := dc.LoadFontFace("./ttf/zh-cn.ttf", 30); err != nil { // 从本地加载字体文件
		panic(err)
	}
	dc.SetRGB(0, 0, 0) // 设置画笔颜色为黑色

	dc.DrawString("谱面ID:"+recent.BeatmapId, 2, 30*1)
	dc.DrawString("玩家名: "+OsuId, 2, 30*2)
	dc.DrawString("玩家ID:"+recent.UserId, 2, 30*3)
	dc.DrawString("游玩模式: "+Model, 2, 30*4)
	dc.DrawString("评价:"+recent.Rank, 2, 30*5)
	dc.DrawString("最大连击 : "+recent.MaxCombo, 2, 30*6)
	dc.DrawString("分数:"+recent.Score, 2, 30*7)
	dc.DrawString("游玩时间: ", 2, 30*8)
	dc.DrawString(recent.Date, 2, 30*9)
	//dc.DrawString("Count300: "+recent.Count300, 2, 30*9)
	//dc.DrawString("Count100: "+recent.Count100, 2, 30*10)
	//dc.DrawString("Count50: "+recent.Count50, 2, 30*11)
	//dc.DrawString("CountMiss: "+recent.CountMiss, 2, 30*12)

	TIME := time.Now().String()

	comma2 := strings.Index(TIME, " ")
	TIME = TIME[:comma2]

	return dc.GenMessageAuto()

}

/*[
	{
		"beatmap_id":"1480628",

		"score":"681257",

		"maxcombo":"353",

		"count50":"1",

		"count100":"23",

		"count300":"385",

		"countmiss":"17",

		"countkatu":"129",

		"countgeki":"436",

		"perfect":"0",

		"enabled_mods":"0",

		"user_id":"18141351",

		"date":"2022-06-11 07:11:21",

		"rank":"A",

		"score_id":"496087202"
	}
]
*/
