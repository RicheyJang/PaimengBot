package pixiv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "好康的",
	Usage: `用法：
	美图/涩图 [Tag]* [数量num]?：num(默认1张)张随机Pixiv美图，来自经过筛选的图库
示例：
	美图 胡桃 2：丢给你两张精选胡桃的美(se)图~
	来两张胡桃的涩图：等同于上一条
另外，高级用法询问管理员哦~[dog]`,
	SuperUsage: `特别用法：(在私聊中)
	色图r [Tag]* [数量num]?：你懂得`,
	Classify: "好康的",
}
var proxy *manager.PluginProxy

type PictureInfo struct {
	Title string // 标题
	URL   string // 图片链接

	Tags   []string // 标签
	PID    int64
	P      int    // 分P
	Author string // 作者
	UID    int64  // 作者UID

	Src string // 无需填写，来源图库
}
type PictureGetter func(tags []string, num int, isR18 bool) []PictureInfo

var ( // 若有新的图库加入，修改以下两个Map即可，会自动适配
	getterMap = map[string]PictureGetter{ // 各个图库的取图函数映射
		"lolicon": getPicturesFromLolicon,
		"omega":   getPicturesFromOmega,
	}
	getterScale = map[string]int{ // 从各个图库取图的初始比例
		"lolicon": 5,
		"omega":   0,
	}
)

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"美图", "涩图", "色图", "瑟图"}).SetBlock(true).SecondPriority().Handle(getPictures)
	proxy.OnCommands([]string{"美图r", "涩图r", "色图r", "瑟图r"}).SetBlock(true).SecondPriority().Handle(getPictures)
	proxy.OnRegex(`来?([\d一两二三四五六七八九十]*)[张页点份发](.*)的?[色涩美瑟]图([rR]?)`).SetBlock(true).SetPriority(4).Handle(getPicturesWithRegex)
	proxy.AddConfig("omega.setu", false) // 在请求非R18图片时，是否从Omega图库中拿取nsfw=1(setu)的图片
	proxy.AddAPIConfig(consts.APIOfHibiAPIKey, "api.obfs.dev")
	proxy.AddConfig("proxy", "i.pixiv.re")
	for k, v := range getterScale { // 各个图库取图比例配置
		proxy.AddConfig(fmt.Sprintf("scale.%s", k), v)
	}
}

// 消息处理函数 -----

func getPictures(ctx *zero.Ctx) {
	// 命令
	isR := false
	cmd := utils.GetCommand(ctx)
	if strings.HasSuffix(cmd, "r") || strings.HasSuffix(cmd, "R") {
		if !utils.IsMessagePrimary(ctx) {
			ctx.Send("滚滚滚")
			return
		}
		isR = true
	}
	// 参数
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	args := strings.Split(arg, " ")
	num := getCmdNum(args[len(args)-1])
	if num > 1 {
		args = args[:len(args)-1]
	}
	// 发图
	newDownloader(args, num, isR).send(ctx)
}

func getPicturesWithRegex(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) <= 3 { // 正则出错
		ctx.Send("？")
		return
	}
	num := getCmdNum(subs[1])
	tags := strings.Split(subs[2], " ")
	isR := false
	if len(subs[3]) > 0 {
		if !utils.IsMessagePrimary(ctx) {
			ctx.Send("滚滚滚")
			return
		}
		isR = true
	}
	// 发图
	newDownloader(tags, num, isR).send(ctx)
}

func getCmdNum(num string) int {
	if r, ok := chineseNumToInt[num]; ok {
		return r
	}
	r, err := strconv.Atoi(num)
	if err != nil || r <= 0 {
		return 1
	}
	return r
}

var chineseNumToInt = map[string]int{
	"一": 1,
	"两": 2,
	"二": 2,
	"三": 3,
	"四": 4,
	"五": 5,
	"六": 6,
	"七": 7,
	"八": 8,
	"九": 9,
	"十": 10,
}
