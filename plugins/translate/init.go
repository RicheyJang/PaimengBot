package translate

import (
	"regexp"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "翻译",
	Usage: `用法：
	翻译 [XXX]：将任意语种内容XXX翻译成中文
	英语 [XXX]：将任意语种内容XXX翻译成英语
	日语 [XXX]：将任意语种内容XXX翻译成日语
此外，若管理员开启了百度API，可以做到例如：
	文言文翻译成西语 尔食之：将文言文"尔食之"翻译成西班牙语
	（"翻译成"的左右两侧可以是任意常见语种`,
	SuperUsage: `默认使用的是有道Hook，翻译效果可能不大好，且只支持少数语种
想要更好的翻译效果，可以在配置中配置上baidu.appid和baidu.key来使用百度翻译
appid和key可以在https://api.fanyi.baidu.com/上注册获取，完全免费（百度打钱
	baidu.maxdaily: 每日最大请求字符上限，防止超限
	baidu.max: 单次翻译的最大字符数`,
	Classify: "实用工具",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"翻译 "}).SetBlock(true).ThirdPriority().Handle(genTranslateHandler("auto", "zh"))
	proxy.OnCommands([]string{"英语"}).SetBlock(true).ThirdPriority().Handle(genTranslateHandler("auto", "en"))
	proxy.OnCommands([]string{"日语"}).SetBlock(true).ThirdPriority().Handle(genTranslateHandler("auto", "jp"))
	proxy.OnRegex("(.*)翻译成?(\\S+)\\s+(.*)", zero.OnlyToMe).SetBlock(true).SetPriority(4).Handle(regexHandler)
	_, _ = proxy.AddScheduleDailyFunc(0, 1, initialBaiduDailyCount)
	proxy.AddConfig(consts.PluginConfigCDKey, "3s")
	proxy.AddConfig("max", 150) // 单次翻译语句最大长度
	proxy.AddConfig("baidu.appid", "")
	proxy.AddConfig("baidu.key", "")
	proxy.AddConfig("baidu.maxdaily", 60000) // 百度翻译每日请求字符数上限
}

func genTranslateHandler(from, to string) func(*zero.Ctx) {
	return func(ctx *zero.Ctx) {
		str := strings.TrimSpace(utils.GetArgs(ctx))
		if utils.StringRealLength(str) > int(proxy.GetConfigInt64("max")) {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("你说的太长啦！短一点！"))
			return
		}
		trans, err := Translate(str, from, to)
		if err != nil {
			log.Warnf("翻译失败，err: %v", err)
		}
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(trans))
	}
}

func regexHandler(ctx *zero.Ctx) {
	arg := strings.TrimSpace(ctx.ExtractPlainText())
	reg := regexp.MustCompile("(\\S*)翻译成?(\\S+)\\s+(.*)")
	subs := reg.FindStringSubmatch(arg)
	if len(subs) < 4 {
		ctx.Send("翻译什么？")
		return
	}
	if utils.StringRealLength(subs[3]) > int(proxy.GetConfigInt64("max")) {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("你说的太长啦！短一点！"))
		return
	}
	trans, err := Translate(subs[3], subs[1], subs[2])
	if err != nil {
		log.Warnf("翻译失败，err: %v", err)
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text(trans))
}

// Translate 通用翻译函数，from\to参数为纯小写，使用BaiduTransLangMap中的key或val，返回翻译结果，当err != nil时，返回的是错误原因
func Translate(str, from, to string) (string, error) {
	// 优先尝试百度接口
	baidu, err := BaiduTranslate(str, from, to)
	if err == nil {
		return baidu, nil
	}
	// 失败了则尝试有道
	return FreeTranslate(str, from, to)
}
