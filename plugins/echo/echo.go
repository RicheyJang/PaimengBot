package echo

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		log.Error("echo init fail")
	}
	proxy.OnCommands([]string{"echo"}).SetBlock(true).FirstPriority().Handle(EchoHandler)
	proxy.AddConfig("times", 2)
}

var info = manager.PluginInfo{
	Name: "复读",
	Usage: `
用法：
	echo 复读内容：将echo后的内容进行复读
`,
}

func EchoHandler(ctx *zero.Ctx) {
	str := utils.GetArgs(ctx)
	tm := proxy.GetConfigInt64("times")
	for i := int64(0); i < tm; i++ {
		ctx.Send(str)
	}
}
