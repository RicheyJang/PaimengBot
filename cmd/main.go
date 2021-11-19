package main

import (
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	// 普通插件
	_ "github.com/RicheyJang/PaimengBot/plugins/echo"
)

func main() {
	// 全局初始化工作在manager.init()中进行（包括初始化命令行参数）
	// 启动服务
	zero.RunAndBlock(zero.Config{
		NickName:      []string{viper.GetString("nickname")},
		CommandPrefix: "",
		SuperUsers:    viper.GetStringSlice("superuser"),
		Driver: []zero.Driver{
			driver.NewWebSocketClient(viper.GetString("server"), ""),
		},
	})
}
