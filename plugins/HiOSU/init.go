package HiOSU

import "github.com/RicheyJang/PaimengBot/manager"

var info = manager.PluginInfo{
	Name: "OSU查询",
	Usage: `音游osu!的相关信息查询，使用前请先 "OSU账号绑定" 来绑定OSUid
用法：
	OSU账号绑定 [ID]：绑定OSU账号，后面输入自己的纯数字ID或者你的OSU用户名，不填则取消绑定
	OSU账号查看：查看你绑定的ID
	OSU信息 [模式代号]：查看你的账号信息，模式代号为一个数字，默认为0
	OSU最近成绩 [模式代号] :获取最新一次的成绩
		模式代号 0 = osu!标准, 1 = Taiko, 2 = CtB, 3 = osu!mania`,
	SuperUsage: `config-plugin配置项：
	hiosu.key: 在 https://osu.ppy.sh/p/api 上申请的API KEY，必需`,
	Classify: "游戏查询",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"OSU账号绑定", "osu账号绑定"}).SetBlock(true).Handle(BindOSUidHandler)  //绑定账号
	proxy.OnCommands([]string{"OSU账号查看", "osu账号查看"}).SetBlock(true).Handle(ReferOSUidHandler) //账号查看
	proxy.OnCommands([]string{"OSU信息", "osu信息"}).SetBlock(true).Handle(MineInfoHandler)       //查看自己的账号信息
	proxy.OnCommands([]string{"OSU最近成绩", "osu最近成绩"}).SetBlock(true).Handle(RecentPlayHandler) //查看自己的账号信息
	proxy.AddConfig("key", "")
}
