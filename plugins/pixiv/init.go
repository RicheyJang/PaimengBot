package pixiv

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
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

type pictureGetter func(tags []string, num int, isR18 bool) []string

var ( // 若有新的图库加入，修改以下两个Map即可，会自动适配
	getterMap = map[string]pictureGetter{ // 各个图库的取图函数映射
		"lolicon": getPicturesFromLolicon,
		"omega":   getPicturesFromOmega,
	}
	getterScale = map[string]int{ // 从各个图库取图的比例
		"lolicon": 5,
		"omega":   5,
	}
)

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"美图", "涩图", "色图", "瑟图"}).SetBlock(true).SecondPriority().Handle(getPictures)
	proxy.OnRegex(`来?([\d一两二三四五六七八九十]*)[张页点份发](.*)的?[色涩美瑟]图([rR]?)`).SetBlock(true).SetPriority(4).Handle(getPicturesWithRegex)
	for k, v := range getterScale { // 各个图库取图比例配置
		proxy.AddConfig(fmt.Sprintf("scale.%s", k), v)
	}
}

func getPictures(ctx *zero.Ctx) {
	//arg := strings.TrimSpace(utils.GetArgs(ctx))

}

func getPicturesWithRegex(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) <= 3 { // 正则出错
		ctx.Send("？")
		return
	}

}

// 从各个图库随机获取图片，返回图片URL切片
func getRandomPictures(tags []string, num int, isR18 bool) []string {
	return []string{}
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
