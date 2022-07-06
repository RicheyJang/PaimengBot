package github

import (
	"net/url"
	"strconv"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "GitHub查询",
	Usage: `查询某个Github仓库的相关信息
用法：
	github [仓库名称]
例子：
	github RicheyJang/PaimengBot`,
	SuperUsage: `config-plugin配置项：
	github.maxresult：一次性最多返回多少个仓库信息`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex(`^github\s+([\x00-\xff]+)$`).SetBlock(true).SetPriority(4).Handle(handleReg)
	proxy.AddConfig("maxResult", 1)
}

const githubAPI = "https://api.github.com/search/repositories"

func handleReg(ctx *zero.Ctx) {
	// 最大结果数
	var maxResult = int(proxy.GetConfigInt64("maxResult"))
	if maxResult <= 0 {
		maxResult = 1
	}
	// 调用API
	var c = client.NewHttpClient(nil)
	target := url.QueryEscape(utils.GetRegexpMatched(ctx)[1])
	// 获取结果
	json, errApi := c.GetGJson(githubAPI + "?q=" + target + "&per_page=" + strconv.FormatInt(int64(maxResult), 10))
	if errApi != nil {
		log.Errorf("GitHubApi返回解析错误,err=%s", errApi.Error())
		ctx.Send("失败了...")
		return
	}
	if count := json.Get("total_count").Int(); count <= 0 {
		ctx.Send("没有找到这样的仓库~")
		return
	}

	var names []string
	for i, result := range json.Get("items").Array() {
		if i >= maxResult {
			break
		}
		names = append(names, result.Get("full_name").String())
	}
	sendImageChain(ctx, names)
}

func sendImageChain(ctx *zero.Ctx, fullNames []string) {
	for _, name := range fullNames {
		ctx.Send(message.Image("https://opengraph.githubassets.com/0/" + name))
	}
}
