package genshin_query

import (
	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "原神便签查询",
	Usage: `需要预先绑定cookie和uid，参见：帮助 米游社管理
用法：
	原神体力：即可查询绑定的原神角色当前树脂、宝钱、派遣等信息
	原神便签：同上`,
	SuperUsage: `config-plugin配置项：
	genshin_query.left: 体力恢复完成时间展示格式
		true则显示剩余时间，false则显示时间点`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnFullMatch([]string{"原神体力", "原神便签", "原神树脂"}).SetBlock(true).SetPriority(3).Handle(queryInfo)
	proxy.AddConfig("left", false)
}

func queryInfo(ctx *zero.Ctx) {
	// 查询绑定
	userUid, userCookie, cookieMsg, err := mihoyo.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookieMsg))
		return
	}
	// 请求便签
	msg, _, err := Query(userUid, userCookie, proxy.GetConfigBool("left"))
	if err != nil {
		log.Errorf("Query err: %v", err)
	}
	ctx.Send(msg)
}
