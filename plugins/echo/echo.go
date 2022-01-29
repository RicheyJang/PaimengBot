package echo

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{ // [1] 声明插件信息结构变量
	Name: "复读",
	Usage: `
用法：
	echo [复读内容]：将echo后的内容进行复读
`,
}
var proxy *manager.PluginProxy // [2] 声明插件代理变量

func init() {
	proxy = manager.RegisterPlugin(info) // [3] 使用插件信息初始化插件代理
	if proxy == nil {                    // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	proxy.OnCommands([]string{"复读", "echo"}).SetBlock(true).FirstPriority().Handle(EchoHandler) // [4] 注册事件处理函数
	proxy.AddConfig("times", 2)                                                                 // proxy提供的统一配置项管理功能，此函数新增一个配置项times，默认值为2
}

// EchoHandler [5] Handler实现
func EchoHandler(ctx *zero.Ctx) {
	str := utils.GetArgs(ctx)           // 派蒙Bot提供的工具函数，用于获取此次事件的消息参数内容
	tm := proxy.GetConfigInt64("times") // proxy提供的统一配置项管理功能，此函数用于获取int64类型的times配置项值
	for i := int64(0); i < tm; i++ {
		ctx.Send(str)
	}
}
