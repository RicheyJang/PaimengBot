package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	"github.com/RicheyJang/PaimengBot/manager"

	// 普通插件
	_ "github.com/RicheyJang/PaimengBot/plugins/echo"
)

func main() {
	// 全局初始化工作在manager.init()中进行（包括初始化命令行参数）
	// 读取插件配置
	err := manager.FlushConfig(".", "config-plugins.yaml")
	if err != nil {
		log.Error("FlushConfig err: ", err)
		return
	}
	// 初始化数据库
	dbV := viper.Sub("db")
	dbC := new(manager.DBConfig)
	err = dbV.Unmarshal(dbC)
	if err != nil {
		log.Error("Unmarshal DB Config err: ", err)
		return
	}
	err = manager.SetupDatabase(dbV.GetString("type"), *dbC)
	if err != nil {
		log.Error("SetupDatabase err: ", err)
		return
	}
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
