package genshin_sign

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_client"
)

var info = manager.PluginInfo{
	Name: "米游社签到",
	Usage: `需要预先绑定cookie和uid，参见：帮助 米游社管理
用法：
	米游社签到：顾 名 思 义
	米游社定时签到 [打开/关闭]：即可打开/关闭米游社自动定时签到
	米游社签到查询：看看机器人今天有没有帮你签到`,
	SuperUsage: `config-plugin配置项：
	genshin_sign.daily.hour: 每天几点自动签到
	genshin_sign.daily.min: 上述钟点的第几分钟自动签到`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"米游社签到"}).SetBlock(true).SetPriority(3).Handle(singleSignHandler)
	proxy.OnCommands([]string{"米游社自动签到", "米游社定时签到"}).SetBlock(true).SetPriority(3).Handle(autoSignHandler)
	proxy.OnCommands([]string{"米游社查询签到", "查询米游社签到", "米游社签到查询"}).SetBlock(true).SetPriority(3).Handle(querySignHandler)
	proxy.AddConfig("daily.hour", 9)
	proxy.AddConfig("daily.min", 0)
	// 添加定时签到任务
	task_id, _ = proxy.AddScheduleDailyFunc(
		int(proxy.GetConfigInt64("daily.hour")),
		int(proxy.GetConfigInt64("daily.min")),
		auto_sign)
	manager.WhenConfigFileChange(configReload)
}

func Sign(uid string, cookie string) (string, error) {

	g := sign_client.NewGenshinClient()

	gameRolesList := g.GetUserGameRoles(cookie)

	for j := 0; j < len(gameRolesList); j++ {
		//time.Sleep(10 * time.Second)
		msg := ""
		if g.Sign(cookie, gameRolesList[j]) {
			//time.Sleep(10 * time.Second)
			data := g.GetSignStateInfo(cookie, gameRolesList[j])
			msg = fmt.Sprintf("UID:%v, 昵称:%v, 连续签到天数:%v. 签到成功.",
				gameRolesList[j].UID, gameRolesList[j].Name, data.TotalSignDay)
		} else {
			msg = fmt.Sprintf("UID:%v, 昵称:%v. 签到失败.",
				gameRolesList[j].UID, gameRolesList[j].Name)
		}
		return msg, nil
	}
	return "未知错误", nil

}
