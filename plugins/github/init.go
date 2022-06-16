package github

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "GitHub查询",
	Usage: `
查询某个仓库
用法：
	github[查询名称]
例子：
	github RicheyJang/PaimengBot
`,
	Classify: "一般功能",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex("^github\\s*(\\S+)$").SetBlock(true).SetPriority(4).Handle(handleReg)
}

func handleReg(ctx *zero.Ctx) {
	name := utils.GetRegexpMatched(ctx)[1]

	log.Info(name)
}
