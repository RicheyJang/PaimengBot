package geng

import (
	"fmt"
	"strconv"
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
	Name: "搜梗",
	Usage: `
搜索一个梗背后的梗知识
用法：
	搜梗 [梗]+
`,
	SuperUsage: `config-plugin配置项：
	geng.max: 单个梗查询的最大答案条数
	geng.shield: 群聊中答案的屏蔽词列表`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"搜梗", "梗知识"}).SetBlock(true).SecondPriority().Handle(searchGeng)
	proxy.AddConfig("max", 4)
	proxy.AddConfig("shield", []string{"精", "性"}) // 群聊屏蔽词列表
}

const gzsAPI = "https://api.iyk0.com/gzs"

func searchGeng(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if len(arg) == 0 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("参数不对哦，可以看看帮助"))
		return
	}
	// 调用
	url := fmt.Sprintf("%s?msg=%v", gzsAPI, arg)
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(url)
	if err != nil {
		log.Errorf("call API Error: err=%v", err)
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		return
	}
	// 解析
	data := rsp.Get("data").Array()
	if rsp.Get("sum").Int() == 0 || len(data) == 0 {
		log.Errorf("searchGeng failed, %v msg=%v", gzsAPI, rsp.Get("msg"))
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("%v也不知道", utils.GetBotNickname())))
		return
	}
	max := proxy.GetConfigInt64("max")
	str := "可能是："
	for i := 0; i < len(data) && i < int(max); i++ {
		str += "\n" + strconv.FormatInt(int64(i+1), 10) + "、" + data[i].Get("title").String()
	}
	// 消息处理
	str = strings.ReplaceAll(str, "​", "")
	str = strings.ReplaceAll(str, "‌", "")
	if utils.IsMessageGroup(ctx) { // 在群聊中，将屏蔽词替换为*
		shields := proxy.GetConfigStrings("shield")
		for _, shield := range shields {
			str = strings.ReplaceAll(str, shield, "*")
		}
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text(str))
}
