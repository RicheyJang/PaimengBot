package genshin_record

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"image"
	"os"
	"strconv"
)

const GenshinCardDir = consts.GenshinImageDir + "/card" //原神角色卡片位置
const GenshinWorldICONPicDir = consts.GenshinImageDir + "/worldicons"
const GenshinWorldBackgroundPicDir = consts.GenshinImageDir + "/worldbackground"
const GenshinHomeworldPicDir = consts.GenshinImageDir + "/homeworld"

var info = manager.PluginInfo{
	Name: "原神战绩",
	Usage: `
	原神战绩：即可查询原神战绩 ，使用前需要使用“米游社管理”绑定UID`,
	Classify: "原神相关",
}

/**
https://api.daidr.me/#/apis  #API文档
*/

type GenShinInfo struct {
	Retcode int    `json:"retcode"` //状态码 1008:用户信息不匹配  2000:uid参数缺失
	Message string `json:"message"` //状态信息
	Data    struct {
		Role struct {
			AvatarUrl string `json:"AvatarUrl"` //无返回,未知
			Nickname  string `json:"nickname"`  // 角色名称
			Region    string `json:"region"`    //服务器位置
			Level     int    `json:"level"`     //等级
		} `json:"role"`
		Avatars []struct {
			Id                      int    `json:"id"`                        //角色ID
			Image                   string `json:"image"`                     //角色头像图片
			Name                    string `json:"name"`                      //角色名称
			Element                 string `json:"element"`                   //元素
			Fetter                  int    `json:"fetter"`                    //好感度
			Level                   int    `json:"level"`                     //角色等级
			Rarity                  int    `json:"rarity"`                    //星级
			ActivedConstellationNum int    `json:"actived_constellation_num"` //命座数
			CardImage               string `json:"card_image"`                //角色卡片名称
			IsChosen                bool   `json:"is_chosen"`                 //未知
		} `json:"avatars"`
		Stats struct {
			ActiveDayNumber      int    `json:"active_day_number"`      //活跃天数
			AchievementNumber    int    `json:"achievement_number"`     //达成成就数
			AnemoculusNumber     int    `json:"anemoculus_number"`      //风神瞳数量
			GeoculusNumber       int    `json:"geoculus_number"`        //岩神瞳数量
			AvatarNumber         int    `json:"avatar_number"`          //拥有角色数量
			WayPointNumber       int    `json:"way_point_number"`       //解锁锚点数
			DomainNumber         int    `json:"domain_number"`          //解锁秘境数
			SpiralAbyss          string `json:"spiral_abyss"`           //深境螺旋层数
			PreciousChestNumber  int    `json:"precious_chest_number"`  //珍贵宝箱数
			LuxuriousChestNumber int    `json:"luxurious_chest_number"` //华丽宝箱数
			ExquisiteChestNumber int    `json:"exquisite_chest_number"` //精致宝箱数
			CommonChestNumber    int    `json:"common_chest_number"`    //普通宝箱数
			ElectroculusNumber   int    `json:"electroculus_number"`    //雷神瞳数
			MagicChestNumber     int    `json:"magic_chest_number"`     //奇馈宝箱数
			DendroculusNumber    int    `json:"dendroculus_number"`     //风神瞳数量
		} `json:"stats"`
		CityExplorations  []interface{} `json:"city_explorations"` // 未知
		WorldExplorations []struct {
			Level                 int    `json:"level"`                  //声望等级
			ExplorationPercentage int    `json:"exploration_percentage"` //
			Icon                  string `json:"icon"`                   //地区图标
			Name                  string `json:"name"`                   //地区名称
			Type                  string `json:"type"`                   //类型 具体未知
			Offerings             []struct {
				Name  string `json:"name"`  //地区贡献名称 例如忍冬之树
				Level int    `json:"level"` //等级
				Icon  string `json:"icon"`  //图标
			} `json:"offerings"`
			Id              int    `json:"id"`               //
			ParentId        int    `json:"parent_id"`        //
			MapUrl          string `json:"map_url"`          //
			StrategyUrl     string `json:"strategy_url"`     //
			BackgroundImage string `json:"background_image"` //地区背景图片
			InnerIcon       string `json:"inner_icon"`       //地区图标(深色)
			Cover           string `json:"cover"`            //
		} `json:"world_explorations"`
		Homes []struct {
			Level            int    `json:"level"`              // 尘歌壶等级
			VisitNum         int    `json:"visit_num"`          //来访人数
			ComfortNum       int    `json:"comfort_num"`        //舒适度
			ItemNum          int    `json:"item_num"`           //放置物品树
			Name             string `json:"name"`               //地区名称
			Icon             string `json:"icon"`               //图标
			ComfortLevelName string `json:"comfort_level_name"` //舒适度名
			ComfortLevelIcon string `json:"comfort_level_icon"` //舒适度图标
		} `json:"homes"`
	} `json:"data"`
}

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"原神战绩"}).SetBlock(true).Handle(getRecord) //绑定账号
}

func getRecord(ctx *zero.Ctx) {
	UID := getUserUid(ctx.Event.UserID)
	UID = UID[0:len(UID)]

	ServerNum := "0"
	API := "https://api.daidr.me/apis/genshinUserinfo?uid=" + UID + "&server=" + ServerNum

	Info, err := GetInfo(API)
	if err != nil {
		log.Errorf("GetInfo Err: %v", err)
	}

	switch Info.Retcode {

	case 2000:
		log.Errorf(Info.Message)
		ctx.Send(Info.Message + ",请使用 [米游社管理] 绑定UID")
		break

	case 1008:
		ServerNum = "1"
		API = "https://api.daidr.me/apis/genshinUserinfo?uid=" + UID + "&server=" + ServerNum
		Info, _ = GetInfo(API)
		if Info.Retcode == 10101 {
			log.Errorf("GenshinRecord Err:达到原神战绩API日请求次数限制")
			log.Errorf("GenshinRecord Message(10101):  " + Info.Message)
			log.Errorf("GenshinRecord Retcode(10101):" + strconv.Itoa(Info.Retcode))
			ctx.Send("请求次数太多了,明天试试把!")
		} else if Info.Retcode == 0 {
			log.Errorf("获取原神战绩成功")
			Image, _ := getRecordImage(Info, UID)
			ctx.Send(Image)
		} else {
			log.Errorf("GenshinRecord Err:???")
			log.Errorf("GenshinRecord Message:  " + Info.Message)
			log.Errorf("GenshinRecord Retcode " + strconv.Itoa(Info.Retcode))
		}

		break

	case 0:
		log.Errorf("获取原神战绩成功")
		Image, _ := getRecordImage(Info, UID)
		ctx.Send(Image)
		break

	case 10101:
		log.Errorf("GenshinRecord Err:达到原神战绩API日请求次数限制")
		log.Errorf("GenshinRecord Message(10101):  " + Info.Message)
		log.Errorf("GenshinRecord Retcode:" + strconv.Itoa(Info.Retcode))
		ctx.Send("请求次数太多了,明天试试把!")
		break

	default:
		log.Errorf("GenshinRecord Err:???")
		log.Errorf("GenshinRecord Message:  " + Info.Message)
		log.Errorf("GenshinRecord Retcode " + strconv.Itoa(Info.Retcode))
		ctx.Send("发生未知错误!!!")
		break

	}

}

func getRecordImage(GenShin GenShinInfo, UID string) (message.MessageSegment, error) {

	RecordImage := images.NewImageCtxWithBGColor(1000, 5000, "#363839")

	//背景图和Title图
	//ImgBackground, _ := gg.LoadImage("genshin/background/Background1.png") //背景图
	ImgTitle, _ := manager.DecodeStaticImage("genshin/title/Title.png") //Title图
	ImgPlayer, _ := manager.DecodeStaticImage("genshin/Player/Player.png")
	//RecordImage.DrawImage(ImgBackground, 0, 450)
	RecordImage.DrawImage(ImgTitle, 0, 0)
	RecordImage.DrawImage(ImgPlayer, 70, 550)

	/************************基本信息*************************************/

	Nickname := GenShin.Data.Role.Nickname             //玩家名
	Server := getUserServer(GenShin.Data.Role.Region)  //所属服务器
	UserLevel := strconv.Itoa(GenShin.Data.Role.Level) //玩家等级

	RecordImage.UseDefaultFont(25)
	RecordImage.SetHexColor("#e5e5e5")

	ImgName, _ := manager.DecodeStaticImage("genshin/module/module01.png")
	nickname := "昵称"
	RecordImage.DrawString(nickname, 400, 640) //昵称
	RecordImage.DrawImage(ImgName, 395, 650)   //昵称组件框

	ImgLevel, _ := manager.DecodeStaticImage("genshin/module/module03.png")
	level := "等级"
	RecordImage.DrawString(level, 400, 715)   //等级
	RecordImage.DrawImage(ImgLevel, 395, 725) //等级组件框

	ImgServer, _ := manager.DecodeStaticImage("genshin/module/module04.png")
	server := "所属服务器"
	RecordImage.DrawString(server, 515, 715)   //服务器
	RecordImage.DrawImage(ImgServer, 510, 725) //服务器组件框

	RecordImage.UseDefaultFont(30)
	RecordImage.DrawString(Nickname, 410, 680)
	RecordImage.DrawString(Server, 515, 755)
	RecordImage.DrawString(UserLevel, 400, 755)

	RecordImage.UseDefaultFont(30)

	RecordImage.DrawString("UID: "+UID, 400, 800)

	/*******************************角色信息**********************************/

	RecordImage.UseDefaultFont(50)
	Characterinfo := "角色信息"
	var info_Y1 float64 = 880
	RecordImage.DrawString(Characterinfo, 100, info_Y1)
	RecordImage.SetLineWidth(5)
	var info_Y2 = info_Y1 + 30
	RecordImage.DrawLine(85, info_Y2, 900, info_Y2)
	RecordImage.StrokePreserve()

	err := updateCharacterPicture(GenShin)
	if err != nil {
		fmt.Println(err)
	}

	//角色信息
	RecordImage.UseDefaultFont(30)
	//放置角色图片
	for i := 0; i < len(GenShin.Data.Avatars); i++ {

		if i < 4 {
			//左列角色信息
			CharacterPic, _ := gg.LoadImage(GenshinCardDir + "/" + GenShin.Data.Avatars[i].Name + ".png") //角色图片
			ElementPic, _ := getElementPicture(GenShin.Data.Avatars[i].Element)                           //元素图

			//角色信息
			CharacterName := GenShin.Data.Avatars[i].Name
			CharacterLevel := "Lv. " + strconv.Itoa(GenShin.Data.Avatars[i].Level)
			CharacterFetter := "好感度等级: " + strconv.Itoa(GenShin.Data.Avatars[i].Fetter)
			CharacterActivedConstellationNum := "命座数: " + strconv.Itoa(GenShin.Data.Avatars[i].ActivedConstellationNum)

			//贴角色图片
			RecordImage.DrawImage(CharacterPic, 50, 925+(256*i))
			RecordImage.DrawImage(ElementPic, 290, 975+(256*i))

			//贴角色信息文字
			RecordImage.DrawString(CharacterName, 345, float64(1010+(256*i)))
			RecordImage.DrawString(CharacterLevel, 300, float64(1060+(256*i)))
			RecordImage.DrawString(CharacterFetter, 300, float64(1110+(256*i)))
			RecordImage.DrawString(CharacterActivedConstellationNum, 300, float64(1160+(256*i)))
		} else {
			p := i - 4

			//右列
			CharacterPic, _ := gg.LoadImage(GenshinCardDir + "/" + GenShin.Data.Avatars[i].Name + ".png")
			ElementPic, _ := getElementPicture(GenShin.Data.Avatars[i].Element)

			CharacterName := GenShin.Data.Avatars[i].Name
			CharacterLevel := "Lv. " + strconv.Itoa(GenShin.Data.Avatars[i].Level)
			CharacterFetter := "好感度等级: " + strconv.Itoa(GenShin.Data.Avatars[i].Fetter)
			CharacterActivedConstellationNum := "命座数: " + strconv.Itoa(GenShin.Data.Avatars[i].ActivedConstellationNum)

			RecordImage.DrawImage(CharacterPic, 495, 925+(256*p))
			RecordImage.DrawImage(ElementPic, 725, 975+(256*p))

			RecordImage.DrawString(CharacterName, 775, float64(1010+(256*p)))
			RecordImage.DrawString(CharacterLevel, 735, float64(1060+(256*p)))
			RecordImage.DrawString(CharacterFetter, 735, float64(1110+(256*p)))
			RecordImage.DrawString(CharacterActivedConstellationNum, 735, float64(1160+(256*p)))
		}

	}

	/********************************数据总览***********************************/

	RecordImage.UseDefaultFont(50)
	AllInfo := "数据总览"
	var Allinfo_Y1 float64 = 2020
	RecordImage.DrawString(AllInfo, 100, Allinfo_Y1)
	RecordImage.SetLineWidth(5)
	var Allinfo_Y2 = Allinfo_Y1 + 30
	RecordImage.DrawLine(85, Allinfo_Y2, 900, Allinfo_Y2)
	RecordImage.StrokePreserve()
	RecordImage.UseDefaultFont(30)

	var list1 float64 = 100  //第一列X坐标
	var list2 float64 = 500  //第二列X坐标
	var All_Y float64 = 2100 //Y坐标
	var linenum float64 = 0  //列数

	ActiveDayNum := "活跃天数: " + strconv.Itoa(GenShin.Data.Stats.ActiveDayNumber)       //活跃天数
	AchievementNum := "成就达成数: " + strconv.Itoa(GenShin.Data.Stats.AchievementNumber) //成就数
	RecordImage.DrawString(ActiveDayNum, list1, All_Y+50*linenum)
	RecordImage.DrawString(AchievementNum, list2, All_Y+50*linenum)
	linenum++

	//各种神瞳数量
	AnemoculusNum := "风神瞳数: " + strconv.Itoa(GenShin.Data.Stats.AnemoculusNumber)     //风
	GeoculusNum := "风神瞳数: " + strconv.Itoa(GenShin.Data.Stats.GeoculusNumber)         //岩
	ElectroculusNum := "风神瞳数: " + strconv.Itoa(GenShin.Data.Stats.ElectroculusNumber) //雷
	DendroculusNum := "风神瞳数: " + strconv.Itoa(GenShin.Data.Stats.DendroculusNumber)   //草
	RecordImage.DrawString(AnemoculusNum, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(GeoculusNum, list2, (All_Y + 50*linenum))
	linenum++
	RecordImage.DrawString(ElectroculusNum, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(DendroculusNum, list2, (All_Y + 50*linenum))
	linenum++

	AvatarNumber := "角色数: " + strconv.Itoa(GenShin.Data.Stats.AvatarNumber)    //角色数
	WayPointNum := "传送点数: " + strconv.Itoa(GenShin.Data.Stats.WayPointNumber) //传送点数量
	DomainNum := "秘境数: " + strconv.Itoa(GenShin.Data.Stats.DomainNumber)       //秘境数量
	SpiralAbyss := "深境螺旋层数: " + GenShin.Data.Stats.SpiralAbyss              //深境螺旋层数
	RecordImage.DrawString(AvatarNumber, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(WayPointNum, list2, (All_Y + 50*linenum))
	linenum++
	RecordImage.DrawString(DomainNum, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(SpiralAbyss, list2, (All_Y + 50*linenum))
	linenum++

	PreciousChestNum := "珍贵宝箱数: " + strconv.Itoa(GenShin.Data.Stats.PreciousChestNumber)   //珍贵宝箱数
	LuxuriousChestNum := "华丽宝箱数: " + strconv.Itoa(GenShin.Data.Stats.LuxuriousChestNumber) //华丽宝箱数
	ExquisiteChestNum := "精致宝箱数: " + strconv.Itoa(GenShin.Data.Stats.ExquisiteChestNumber) //精致宝箱数
	CommonChestNum := "普通宝箱数: " + strconv.Itoa(GenShin.Data.Stats.CommonChestNumber)       //普通宝箱数
	MagicChestNum := "奇馈宝箱数: " + strconv.Itoa(GenShin.Data.Stats.MagicChestNumber)         //奇馈宝箱数
	RecordImage.DrawString(PreciousChestNum, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(LuxuriousChestNum, list2, (All_Y + 50*linenum))
	linenum++
	RecordImage.DrawString(ExquisiteChestNum, list1, (All_Y + 50*linenum))
	RecordImage.DrawString(CommonChestNum, list2, (All_Y + 50*linenum))
	linenum++
	RecordImage.DrawString(MagicChestNum, list1, (All_Y + 50*linenum))

	/********************************世界探索***********************************/

	//这里计划写一个Goroutine
	updateWorldICONPicture(GenShin)
	updateWorldBackgroundPicture(GenShin)
	updateWorldOfferingsPicture(GenShin)

	RecordImage.UseDefaultFont(50) //这里应该改为usedefeafont
	WorldExploration := "世界探索"
	var WE_Y1 float64 = 2560
	RecordImage.DrawString(WorldExploration, 100, WE_Y1)
	RecordImage.SetLineWidth(5)
	var WE_Y2 = WE_Y1 + 30
	RecordImage.DrawLine(85, WE_Y2, 900, WE_Y2)
	RecordImage.StrokePreserve()
	RecordImage.UseDefaultFont(30)

	var WEP_Y0 int = 2640
	var WEP_Y1 int = WEP_Y0 - 30
	var WEP_X int = 120

	//var HWs_Y float64 //自增变量
	for x := 0; x < len(GenShin.Data.WorldExplorations); x++ {

		WEIconPic, _ := gg.LoadImage(GenshinWorldICONPicDir + "/" + GenShin.Data.WorldExplorations[x].Name + ".png")
		WEBackgroundPic, _ := gg.LoadImage(GenshinWorldBackgroundPicDir + "/" + GenShin.Data.WorldExplorations[x].Name + "BG.png")

		WE_Level := "声望等级: " + strconv.Itoa(GenShin.Data.WorldExplorations[x].Level)
		WE_Exploration := GenShin.Data.WorldExplorations[x].ExplorationPercentage / 10
		ExplorationPercentage := "探索度: " + strconv.Itoa(WE_Exploration) + "%"

		var WE_N_Y int = 2700
		var WE_N_X int = 270
		//RecordImage.LoadFontFace("./ttf/zh-cn.ttf", 50)
		WorldName := GenShin.Data.WorldExplorations[x].Name
		RecordImage.DrawString(WorldName, float64(WE_N_X), float64(WE_N_Y+(180*x)))

		if len(GenShin.Data.WorldExplorations[x].Offerings) != 0 {
			RecordImage.LoadFontFace("./ttf/zh-cn.ttf", 30)
			OfferingPic, _ := gg.LoadImage(GenshinWorldBackgroundPicDir + "/offerings" + "/" + GenShin.Data.WorldExplorations[x].Offerings[0].Name + "Off.png")

			OfferimgLevel := strconv.Itoa(GenShin.Data.WorldExplorations[x].Offerings[0].Level)
			OfferingName := GenShin.Data.WorldExplorations[x].Offerings[0].Name + "等级: " + OfferimgLevel

			OffImg_X := WEP_X + 100
			OffImg_Y := WEP_Y0 + 70

			RecordImage.DrawImage(OfferingPic, OffImg_X, OffImg_Y+(180*x))
			RecordImage.DrawString(OfferingName, float64(OffImg_X+50), float64(OffImg_Y+35+(180*x)))

		}

		RecordImage.DrawImage(WEIconPic, WEP_X, WEP_Y0+(180*x))
		RecordImage.DrawImage(WEBackgroundPic, WEP_X+80, WEP_Y1+(180*x))

		WE_E_Y := 2700
		RecordImage.UseDefaultFont(30)
		RecordImage.DrawString(WE_Level, 550, float64(WE_E_Y+(180*x)))
		RecordImage.DrawString(ExplorationPercentage, 550, float64(WE_E_Y+40+(180*x)))

		//HWs_Y = float64(WEP_Y1 + (180 * x))
	}

	/************************************尘歌壶************************************/
	//RecordImage.LoadFontFace("./ttf/zh-cn.ttf", 50) //这里应该改为usedefeafont
	//Homes := "尘歌壶"
	//var Homes_Y1 float64 = HWs_Y + 260
	//RecordImage.DrawString(Homes, 100, Homes_Y1)
	//RecordImage.SetLineWidth(5)
	//var Homes_Y2 = Homes_Y1 + 30
	//RecordImage.DrawLine(85, Homes_Y2, 900, Homes_Y2)
	//RecordImage.StrokePreserve()
	//RecordImage.LoadFontFace("./ttf/zh-cn.ttf", 30)
	//
	//UpdateHomeworldPicture(GenShin)
	//UpdateHomeworldComfortLevelIconPicture(GenShin)
	//
	//////len(GenShin.Data.Homes)
	//for i := 0; i < len(GenShin.Data.Homes); i++ {
	//	HomePic, _ := gg.LoadImage(GenshinHomeworldPicDir + "/" + GenShin.Data.Homes[i].Name + "HW.png")
	//	RecordImage.DrawImage(HomePic, 120, int(Homes_Y1+60)+254*i)
	//
	//}

	return RecordImage.GenMessageAuto()

}

func getUserUid(id int64) (u string) {
	key := fmt.Sprintf("genshin_uid.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return
	}
	_ = json.Unmarshal(v, &u)
	return
}

// 获取尘歌壶舒适度图标
func updateHomeworldComfortLevelIconPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinHomeworldPicDir + "/levelicon")
	if err != nil {
		return err
	}

	for i := 0; i < len(info.Data.Homes); i++ {
		HomeworldComfortLevelName := info.Data.Homes[i].ComfortLevelName
		path := utils.PathJoin(dir, fmt.Sprintf("%vHWC.png", HomeworldComfortLevelName))
		url := info.Data.Homes[i].ComfortLevelIcon
		bool, err := PathExists(path)
		if bool {
			continue
		} else {
			err = client.DownloadToFile(path, url, 2)
			if err != nil {
				return err
			}

		}

	}
	return nil
}

// 获取尘歌壶背景图  计划另外写一个功能实现尘歌壶
func updateHomeworldPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinHomeworldPicDir)
	if err != nil {
		return err
	}
	//请求
	for i := 0; i < len(info.Data.Homes); i++ {
		HomeworldName := info.Data.Homes[i].Name
		path := utils.PathJoin(dir, fmt.Sprintf("%vHW.png", HomeworldName))
		url := info.Data.Homes[i].Icon
		bool, err := PathExists(path)
		if bool {
			continue
		} else {
			err = client.DownloadToFile(path, url, 2)
			if err != nil {
				return err
			}

		}

	}

	return nil

}

// 下载需要的世界背景图
func updateWorldBackgroundPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinWorldBackgroundPicDir)
	if err != nil {
		return err
	}
	//请求
	for i := 0; i < len(info.Data.WorldExplorations); i++ {
		Explorations := info.Data.WorldExplorations[i].Name
		path := utils.PathJoin(dir, fmt.Sprintf("%vBG.png", Explorations))
		url := info.Data.WorldExplorations[i].BackgroundImage
		bool, err := PathExists(path)
		if bool {
			continue
		} else {
			err = client.DownloadToFile(path, url, 2)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

// 其他的世界探索ICON
func updateWorldOfferingsPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinWorldBackgroundPicDir + "/offerings")
	if err != nil {
		return err
	}
	//请求
	for i := 0; i < len(info.Data.WorldExplorations); i++ {
		if len(info.Data.WorldExplorations[i].Offerings) != 0 {
			OfferingName := info.Data.WorldExplorations[i].Offerings[0].Name
			offeringPath := utils.PathJoin(dir, fmt.Sprintf("%vOff.png", OfferingName))
			Offeringurl := info.Data.WorldExplorations[i].Offerings[0].Icon
			bool, err := PathExists(offeringPath)
			if bool {
				continue
			} else {
				err = client.DownloadToFile(offeringPath, Offeringurl, 2)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

// 下载需要的世界ICON
func updateWorldICONPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinWorldICONPicDir)
	if err != nil {
		return err
	}
	//请求
	for i := 0; i < len(info.Data.WorldExplorations); i++ {
		Explorations := info.Data.WorldExplorations[i].Name
		path := utils.PathJoin(dir, fmt.Sprintf("%v.png", Explorations))
		url := info.Data.WorldExplorations[i].Icon
		bool, err := PathExists(path)
		if bool {
			continue
		} else {
			err = client.DownloadToFile(path, url, 2)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

// 下载需要的角色图片
func updateCharacterPicture(info GenShinInfo) error {
	dir, err := utils.MakeDir(GenshinCardDir)
	if err != nil {
		return err
	}
	//请求
	for i := 0; i < len(info.Data.Avatars); i++ {
		name := info.Data.Avatars[i].Name
		path := utils.PathJoin(dir, fmt.Sprintf("%v.png", name))
		url := info.Data.Avatars[i].CardImage //卡片照片
		bool, err := PathExists(path)
		if bool {
			continue
		} else {
			err = client.DownloadToFile(path, url, 2)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

// 获取返回的Json
func GetInfo(API string) (GenShinInfo, error) {
	c := client.NewHttpClient(nil)
	r, err := c.GetReader(API)
	if err != nil {
		return GenShinInfo{}, err
	}
	defer r.Close()
	// 解析JSON
	d := json.NewDecoder(r)
	var GenShinInfo GenShinInfo
	if err = d.Decode(&GenShinInfo); err != nil {
		return GenShinInfo, err
	}
	return GenShinInfo, nil

}

// 将服务器信息转化
func getUserServer(Region string) string {

	switch Region {

	case "cn_qd01":
		return "B服 - 世界树"

	default:
		return "官服 - 天空岛"

	}
}

// 判断是否存在文件
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取元素图片
func getElementPicture(ElementName string) (image.Image, error) {
	switch ElementName {

	case "Cryo": //冰元素
		return manager.DecodeStaticImage("genshin/element/Cryo.png")

	case "Geo": //岩元素
		return manager.DecodeStaticImage("genshin/element/Geo.png")

	case "Anemo": //风元素
		return manager.DecodeStaticImage("genshin/element/Anemo.png")

	case "Dendro": //草元素
		return manager.DecodeStaticImage("genshin/element/Dendro.png")

	case "Electro": //雷元素
		return manager.DecodeStaticImage("genshin/element/Electro.png")

	case "Hydro": //水元素
		return manager.DecodeStaticImage("genshin/element/Hydro.png")

	default: //火元素
		return manager.DecodeStaticImage("genshin/element/Pyro.png")

	}
}
