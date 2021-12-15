package short_url

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "短网址还原",
	Usage: `用法：
	还原短网址 [短网址]：还原一个短链接，支持任意来源
`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"还原短链接", "还原短链", "还原短网址"}).SetBlock(true).SecondPriority().Handle(findShortURL)
}

func findShortURL(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if len(arg) == 0 {
		ctx.Send("网址呢？")
		return
	}
	url := FindShortURLOrg(arg)
	if len(url) == 0 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("%v也不知道", utils.GetBotNickname())))
	} else {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("原链接:\n%v", url)))
	}
}

// FindShortURLOrg 将短链接解析成原始链接，支持任意来源
func FindShortURLOrg(short string) (url string) {
	url = findShortURLOrgFor33hCo(short)
	if len(url) == 0 {
		url = FindShortURLOrgByLocal(short)
	}
	return
}
