package HiOSU

import (
	"encoding/json"
	"fmt"
	"image"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/images"
	log "github.com/sirupsen/logrus"
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
	Model := GetModel(model) //获取游玩模式

	//获取最近游玩成绩API + 其中 k 是 API的key , u 是 查询用户的纯数字ID 或者  ID ， model 是 查询的模式 , limit 是返回的成绩个数
	OsuAPI := "https://osu.ppy.sh/api/get_user_recent?k=" + key + "&u=" + OSUid + "&m=" + model + "&limit=1"
	recent, err := GetRecentPlay(OsuAPI)
	if err != nil {
		ctx.Send("最近没有玩过" + Model + "这个模式哦")
		return
	}
	Image, err := ToImageRecent(recent, Model, OSUid)
	if err != nil {
		log.Errorf("ToImageRecent err: %v", err)
		ctx.Send("谱面ID:" + recent.BeatmapId +
			"\n玩家名: " + OSUid +
			"\n玩家ID:" + recent.UserId +
			"\n游玩模式: " + Model +
			"\n评价:" + recent.Rank +
			"\n最大连击 : " + recent.MaxCombo +
			"\n分数:" + recent.Score +
			"\n游玩时间: " + recent.Date)
		return
	}
	ctx.Send(Image)
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

func ToImageRecent(recent Recent, Model string, OsuId string) (message.MessageSegment, error) { //生成图片

	ModelImage, err := getModelImage(Model)
	if err != nil {
		return message.MessageSegment{}, err
	}

	width := float64(800)
	height := float64(400)
	var dc = images.NewImageCtxWithBGColor(int(width), int(height), "black")

	ResultImage, _ := GetResultImage(recent)                                            //最终成绩图
	ModelImage, _ = getModelImage(Model)                                                //模式图片
	Count300Image, _ := manager.DecodeStaticImage("HiOSU/rank/CountImage/Count300.png") //各种判定图片
	Count100Image, _ := manager.DecodeStaticImage("HiOSU/rank/CountImage/Count100.png")
	Count50Image, _ := manager.DecodeStaticImage("HiOSU/rank/CountImage/Count50.png")
	CountMissImage, _ := manager.DecodeStaticImage("HiOSU/rank/CountImage/CountMiss.png")

	dc.UseDefaultFont(30)
	dc.SetRGB(1, 1, 1) // 设置画笔颜色为黑色

	dc.DrawImage(ResultImage, 490, 0)

	dc.DrawString("User: "+OsuId, 15, 40)
	dc.DrawString("BeatmapId:"+recent.BeatmapId, 15, 80)

	dc.DrawString("Date: "+recent.Date, 15, 120)

	dc.DrawImage(ModelImage, 15, 153)
	dc.DrawString(Model, 50, 170)

	dc.DrawImage(Count300Image, 15, 180)
	dc.DrawString(": "+recent.Count300, 65, 205)
	dc.DrawImage(Count100Image, 190, 180)
	dc.DrawString(": "+recent.Count100, 240, 205)
	dc.DrawImage(Count50Image, 15, 210)
	dc.DrawString(": "+recent.Count50, 65, 235)
	dc.DrawImage(CountMissImage, 190, 210)
	dc.DrawString(": "+recent.CountMiss, 225, 235)

	dc.DrawString("Score: "+recent.Score, 15, 285)

	dc.DrawString("Max Combo: "+recent.MaxCombo, 15, 320)
	dc.DrawString("Rank: "+recent.Rank, 15, 350)

	return dc.GenMessageAuto()
}

func GetResultImage(recent Recent) (image.Image, error) {

	switch recent.Rank {

	case "XH":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-XH_307x400.png")

	case "X":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-X_307x400.png")

	case "SH":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-SH_307x400.png")

	case "S":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-S_307x400.png")

	case "A":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-A_307x400.png")

	case "B":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-B_307x400.png")

	case "C":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-C_307x400.png")

	case "D":
		return manager.DecodeStaticImage("HiOSU/rank/ranking-D_307x400.png")

	default:
		return manager.DecodeStaticImage("HiOSU/rank/raning-default.png")
	}
}
