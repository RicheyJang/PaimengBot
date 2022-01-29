package hhsh

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "缩写翻译",
	Usage: `
将网络上各种纯小写翻译成人话（老年人跟不上了
用法：
	缩写翻译 [全小写缩写]+：将缩写翻译成人话
`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"缩写翻译", "简写翻译"}).SetBlock(true).SecondPriority().Handle(handleTranslate)
	proxy.OnRegex("^什么是([a-z]+)$").SetBlock(true).SetPriority(4).Handle(handleTranslateReg)
	proxy.OnRegex("^([a-z]+)是?什么意思$").SetBlock(true).SetPriority(4).Handle(handleTranslateReg)
	proxy.AddConfig("max", 5)
}

func handleTranslate(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if len(arg) == 0 {
		ctx.Send("？")
		return
	}
	translateHHSH(ctx, arg)
}

func handleTranslateReg(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) <= 1 { // 不予理睬
		return
	}
	translateHHSH(ctx, subs[1])
}

const hhshAPI = "https://lab.magiconch.com/api/nbnhhsh/guess"

func translateHHSH(ctx *zero.Ctx, text string) {
	c := client.NewHttpClient(nil)
	res, err := c.PostJson(hhshAPI, map[string]interface{}{"text": text})
	if err != nil {
		log.Errorf("translateHHSH post err: %v", err)
		ctx.Send("翻译失败了...")
		return
	}
	var str string
	for _, item := range res.Array() { // 所有查询结果
		var line string
		name := item.Get("name").String()
		trans := item.Get("trans").Array()
		for i, tran := range trans { // 每一种翻译
			if i != 0 && i >= int(proxy.GetConfigInt64("max")) {
				break
			}
			if len(line) == 0 {
				line = fmt.Sprintf("%v有可能是：%v", name, tran.String())
			} else {
				line += fmt.Sprintf("、%v", tran.String())
			}
		}
		if len(line) > 0 { // 有结果，拼接结果
			if len(str) > 0 {
				str += "\n"
			}
			str += line
		}
	}
	if len(str) == 0 {
		ctx.Send(fmt.Sprintf("%s也不知道...", utils.GetBotNickname()))
		return
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text(str))
}
