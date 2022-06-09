package jokes

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"strings"
)

var info = manager.PluginInfo{
	Name: "讲个笑话",
	Usage: `用法：
	讲个笑话  : 随机输出一条笑话
	重载笑话  : 重新加载笑话文件 `,
	Classify: "一般功能",
}

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"讲个笑话"}).SetBlock(true).SecondPriority().Handle(getJoke)
	proxy.OnCommands([]string{"重载笑话"}).SetBlock(true).SecondPriority().Handle(reloadFile)
	proxy.AddConfig("jokes", []string{}) // 笑话列表
}

func getJoke(ctx *zero.Ctx) {
	name := ctx.Event.Sender.Name()
	ctx.SendChain(message.Text(getjoke(name)))

}

func reloadFile(*zero.Ctx) {
	LoadJokesFromDir(consts.JokesDir)

}

func getjoke(name string) string {
	l := len(jokesList)
	randJoke := jokesList[rand.Intn(l)]
	return strings.ReplaceAll(randJoke, "%name", name)
}
