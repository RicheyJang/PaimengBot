package echo

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"echo"}).SetBlock(true).FirstPriority().Handle(EchoHandler)
	proxy.AddConfig("times", 2)
	//_, err := proxy.AddScheduleFunc("@every 1s", func() {
	//	log.Info("echo schedule")
	//})
	//if err != nil {
	//	log.Error("echo AddScheduleFunc err:", err)
	//}
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
