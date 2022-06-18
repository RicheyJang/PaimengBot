package HiOSU

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"image"
	"io"
	"strings"
	"time"

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
	PP          string `json:"pp_raw"`          //PP总数
	Accuracy    string `json:"accuracy"`        //准确率
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
		Image, _ := ToImageUser(USER, Model)
		ctx.Send(Image)
	}
}

func GetMyInfo(API string) (User, error) {
	// 调用
	c := client.NewHttpClient(nil)
	r, err := c.GetReader(API)
	if err != nil {
		return User{}, err
	}
	defer func(r io.ReadCloser) {
		err := r.Close()
		if err != nil {
			return
		}
	}(r)
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

func ToImageUser(user User, Model string) (message.MessageSegment, error) { //生成图片(需要修改)
	width := float64(465)
	height := float64(240)
	var dc = images.NewImageCtx(int(width), int(height)) //生成图片大小

	LogoFile, err := manager.GetStaticFile("HiOSU/Logo/Logo_96x97.png") //读取OSU图标
	if err != nil {
		log.Errorf("get HiOSU Logo.png error: %v", err)
		return message.MessageSegment{}, nil
	}
	LogoImage, _, err := image.Decode(LogoFile) // 解码为image.Image
	if err != nil {
		log.Errorf("Decode HiOSU Logo.png error: %v", err)
	}

	//读取各种模式的图标
	StDFile, err := manager.GetStaticFile("HiOSU/Model/std_20x20.png") //Osu!模式图标
	if err != nil {
		log.Errorf("get Osu!Model.png error: %v", err)
	}
	CatchFile, err := manager.GetStaticFile("HiOSU/Model/catch_20x20.png") //Catch模式图标
	if err != nil {
		log.Errorf("get CatchModel.png error: %v", err)
	}
	MainaFile, err := manager.GetStaticFile("HiOSU/Model/mania_20x20.png") //Osu!Maina模式图标
	if err != nil {
		log.Errorf("get MainaModel.png error: %v", err)
	}
	TaikoFile, err := manager.GetStaticFile("HiOSU/Model/taiko_20x20.png") //Taiko模式图标
	if err != nil {
		log.Errorf("get TaikoModel.png error: %v", err)
	}

	StDImage, _, err := image.Decode(StDFile)     // StD图片解码为image.Image
	CtBImage, _, err := image.Decode(CatchFile)   // CtD图片解码为image.Image
	MainaImage, _, err := image.Decode(MainaFile) // Maina图片解码为image.Image
	TaikoImage, _, err := image.Decode(TaikoFile) // Taiko图片解码为image.Image

	dc.SetHexColor("#000000") // 设置画笔颜色为黑
	dc.Clear()                // 使用当前颜色（绿）填满画布，即设置背景色

	err = dc.LoadFontFace("./ttf/zh-cn.ttf", 20)
	if err != nil {
		return message.MessageSegment{}, err
	}
	dc.SetRGB(1, 1, 1) // 设置画笔颜色为白

	dc.DrawImage(LogoImage, 10, 10) //贴OSU图标

	dc.DrawString("Country :"+user.Country, 130, 40)
	err = dc.LoadFontFace("./ttf/zh-cn.ttf", 40) //字体设置大一些
	if err != nil {
		return message.MessageSegment{}, err
	}
	dc.DrawString(user.UserName, 130, 80) //显示UserName
	err = dc.LoadFontFace("./ttf/zh-cn.ttf", 20)
	if err != nil {
		return message.MessageSegment{}, err
	}
	dc.DrawString("Join Date: "+user.JoinDate, 130, 100)
	dc.DrawString("-----------------------------------------------", 0, 123)

	var ModelImage image.Image
	switch Model {

	case "Osu!Mania":
		ModelImage = MainaImage
	case "Taiko":
		ModelImage = TaikoImage
	case "Catch":
		ModelImage = CtBImage
	default:
		ModelImage = StDImage

	}

	dc.DrawImage(ModelImage, 10, 130)
	dc.DrawString(Model, 35, 147)
	UpdateTime := time.Now().String()
	comma := strings.Index(UpdateTime, " ")
	UpdateTime = UpdateTime[:comma]
	dc.DrawString("Update Time :"+UpdateTime, 190, 147)
	dc.DrawString("Global Rank :", 10, 170)
	dc.DrawString("Country Rank  :", 190, 170)
	dc.DrawString("#"+user.GlobalRank, 100, 200)
	dc.DrawString("#"+user.CountryRank, 320, 200)
	dc.DrawString("PP: "+user.PP, 10, 230)
	user.Accuracy = user.Accuracy[:5] //准确度保留后两位
	dc.DrawString("Accuracy: "+user.Accuracy+"%", 190, 230)

	return dc.GenMessageAuto()
}
