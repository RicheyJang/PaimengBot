package main

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	// 基本插件，建议不要删除，可能会造成依赖问题
	_ "github.com/RicheyJang/PaimengBot/basic/auth"
	_ "github.com/RicheyJang/PaimengBot/basic/ban"
	_ "github.com/RicheyJang/PaimengBot/basic/event"
	_ "github.com/RicheyJang/PaimengBot/basic/inspection"
	_ "github.com/RicheyJang/PaimengBot/basic/invite"
	_ "github.com/RicheyJang/PaimengBot/basic/limiter"
	_ "github.com/RicheyJang/PaimengBot/basic/nickname"

	// 普通插件
	_ "github.com/RicheyJang/PaimengBot/plugins/echo"
)

func main() {
	// 全局初始化工作在manager.init()中进行（包括初始化命令行参数）
	// 刷新插件配置（必须在main中进行）
	err := manager.FlushConfig(consts.DefaultConfigDir, consts.PluginConfigFileName)
	if err != nil {
		log.Fatal("FlushConfig err: ", err)
	}
	// 启动服务
	log.Infof("读取超级管理员列表：%v", viper.GetStringSlice("superuser"))
	zero.RunAndBlock(zero.Config{
		NickName:      []string{viper.GetString("nickname")},
		CommandPrefix: "",
		SuperUsers:    viper.GetStringSlice("superuser"),
		Driver: []zero.Driver{
			driver.NewWebSocketClient(viper.GetString("server.address"), viper.GetString("server.token")),
		},
	})
}
