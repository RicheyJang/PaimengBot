package HiOSU

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils/client"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Recent struct {
	BeatmapId string `json:"beatmap_id"` //谱面ID
	Score     string `json:"score"`      //分数
	MaxCombo  string `json:"maxcombo"`   //最大连击
	UserId    string `json:"user_id"`    //玩家ID
	Rank      string `json:"rank"`       //评价
	Count300  string `json:"count300"`   //GREAT数
	Count100  string `json:"count100"`   //GOOD数
	Count50   string `json:"count50"`    //BAD数
	CountMiss string `json:"countmiss"`  //MISS数
	Date      string `json:"date"`       //游玩日期
}

func RecentPlayHandler(ctx *zero.Ctx) {
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
	Model := GetModel(model) //获取游玩模式

	//获取最近游玩成绩API + 其中 k 是 API的key , u 是 查询用户的纯数字ID 或者  ID ， model 是 查询的模式 , limit 是返回的成绩个数
	OsuAPI := "https://osu.ppy.sh/api/get_user_recent?k=51b88dd53687332618935b74d5a3bf22c8326826&u=" + OSUid + "&m=" + model + "&limit=1"
	recent, err := GetRecentPlay(OsuAPI)
	if err != nil {
		ctx.Send("最近没有玩过" + Model + "这个模式哦")
		return
	}
	ctx.Send(ToImageRecent(recent, Model, OSUid))
}

func GetRecentPlay(API string) (Recent, error) {
	// 调用
	c := client.NewHttpClient(nil)
	r, err := c.GetReader(API)
	if err != nil {
		return Recent{}, err
	}
	defer r.Close()
	// 解析
	d := json.NewDecoder(r)
	var recents []Recent
	if err = d.Decode(&recents); err != nil {
		return Recent{}, err
	}
	if len(recents) == 0 {
		return Recent{}, fmt.Errorf("recent info is empty")
	}
	return recents[0], nil
}

func ToImageRecent(recent Recent, Model string, OsuId string) message.MessageSegment { //生成图片(需要修改)
	str := "谱面ID:" + recent.BeatmapId +
		"\n玩家名: " + OsuId +
		"\n玩家ID:" + recent.UserId +
		"\n游玩模式: " + Model +
		"\n评价:" + recent.Rank +
		"\n最大连击 : " + recent.MaxCombo +
		"\n分数:" + recent.Score +
		"\n游玩时间: " + recent.Date
	//"\nCount300: " + recent.Count300 +
	//"\nCount100: " + recent.Count100 +
	//"\nCount50: " + recent.Count50 +
	//"\nCountMiss: " + recent.CountMiss
	return images.GenStringMsg(str)
}
