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
	_ "github.com/RicheyJang/PaimengBot/basic/help"
	_ "github.com/RicheyJang/PaimengBot/basic/invite"
	_ "github.com/RicheyJang/PaimengBot/basic/limiter"
	_ "github.com/RicheyJang/PaimengBot/basic/nickname"
	_ "github.com/RicheyJang/PaimengBot/basic/sc"

	// 普通插件
	_ "github.com/RicheyJang/PaimengBot/plugins/COVID"
	_ "github.com/RicheyJang/PaimengBot/plugins/admin"
	_ "github.com/RicheyJang/PaimengBot/plugins/bilibili"
	_ "github.com/RicheyJang/PaimengBot/plugins/chat"
	_ "github.com/RicheyJang/PaimengBot/plugins/contact"
	_ "github.com/RicheyJang/PaimengBot/plugins/echo"
	_ "github.com/RicheyJang/PaimengBot/plugins/emoji_mix"
	_ "github.com/RicheyJang/PaimengBot/plugins/geng"
	_ "github.com/RicheyJang/PaimengBot/plugins/genshin"
	_ "github.com/RicheyJang/PaimengBot/plugins/hhsh"
	_ "github.com/RicheyJang/PaimengBot/plugins/idioms"
	_ "github.com/RicheyJang/PaimengBot/plugins/inspection"
	_ "github.com/RicheyJang/PaimengBot/plugins/keyword"
	_ "github.com/RicheyJang/PaimengBot/plugins/music"
	_ "github.com/RicheyJang/PaimengBot/plugins/netease"
	_ "github.com/RicheyJang/PaimengBot/plugins/note"
	_ "github.com/RicheyJang/PaimengBot/plugins/pixiv"
	_ "github.com/RicheyJang/PaimengBot/plugins/pixiv_query"
	_ "github.com/RicheyJang/PaimengBot/plugins/pixiv_rank"
	_ "github.com/RicheyJang/PaimengBot/plugins/random"
	_ "github.com/RicheyJang/PaimengBot/plugins/short_url"
	_ "github.com/RicheyJang/PaimengBot/plugins/statistic"
	_ "github.com/RicheyJang/PaimengBot/plugins/translate"
	_ "github.com/RicheyJang/PaimengBot/plugins/weather"
	_ "github.com/RicheyJang/PaimengBot/plugins/welcome"
	_ "github.com/RicheyJang/PaimengBot/plugins/whatanime"
	_ "github.com/RicheyJang/PaimengBot/plugins/withdraw"
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
