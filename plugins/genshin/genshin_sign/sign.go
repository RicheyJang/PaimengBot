package genshin_sign

import (
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_cookie"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "米游社签到",
	Usage: `需要预先绑定cookie和uid，参见：帮助 米游社管理
用法：
	米游社签到：顾 名 思 义
	米游社定时签到 [打开/关闭]：即可打开/关闭米游社自动定时签到，仅限好友私聊
	米游社信息：看看你有没有设置cookie和uid、有没有打开自动定时签到`,
	SuperUsage: `config-plugin配置项：
	genshin_sign.group: 是(true)否(false)允许在群聊中开启米游社自动签到，并向群聊中推送签到信息
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
	proxy.OnFullMatch([]string{"米游社签到"}).SetBlock(true).SetPriority(3).Handle(singleSignHandler)
	proxy.OnFullMatch([]string{"米游社信息", "米游社info"}).SetBlock(true).SetPriority(3).Handle(queryAutoHandler)
	proxy.OnCommands([]string{"米游社定时签到", "米游社自动签到"}, checkCouldGroup).SetBlock(true).SetPriority(3).Handle(autoSignHandler)
	proxy.AddConfig("group", false)
	proxy.AddConfig("daily.hour", 9)
	proxy.AddConfig("daily.min", 0)
	// 添加定时签到任务
	manager.WhenConfigFileChange(configReload)
}

func singleSignHandler(ctx *zero.Ctx) {
	userUid, userCookie, cookieMsg, err := mihoyo.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookieMsg))
		return
	}
	msg, err := Sign(userUid, userCookie)
	if err != nil {
		log.Errorf("Sign(uid=%v) err: %v", userUid, err)
	}
	ctx.Send(images.GenStringMsg(msg))
}

func checkCouldGroup(ctx *zero.Ctx) bool {
	if utils.IsMessagePrimary(ctx) {
		return true
	}
	return proxy.GetConfigBool("group")
}

func autoSignHandler(ctx *zero.Ctx) {
	// 接收参数 判断是开还是关
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if strings.HasPrefix(args, "开") || args == "打开" {
		// 判断是否设置了cookie和uid
		_, _, cookieMsg, err := mihoyo.GetUidCookieById(ctx.Event.UserID)
		if err != nil {
			ctx.Send(images.GenStringMsg(cookieMsg))
			return
		}
		// 添加定时
		ctx.Send(setAutoEvent(true, ctx.Event.GroupID, ctx.Event.UserID))
	} else if strings.HasPrefix(args, "关") {
		ctx.Send(setAutoEvent(false, ctx.Event.GroupID, ctx.Event.UserID))
	} else {
		// 不知道啥情况
		ctx.Send(`？可以看看帮助`)
		return
	}
	return
}

func setAutoEvent(open bool, group int64, id int64) (msg string) {
	event := EventFrom{
		IsFromGroup: group != 0,
		FromId:      strconv.FormatInt(group, 10),
		QQ:          strconv.FormatInt(id, 10),
		Auto:        open,
	}
	if !event.IsFromGroup {
		event.FromId = strconv.FormatInt(id, 10)
	}
	err := PutEventFrom(id, event)
	if err != nil {
		log.Errorf("PutEventFrom err: %v", err)
		return "失败了..."
	}
	if open {
		return "定时签到已打开"
	} else {
		return "定时签到已关闭"
	}
}

func queryAutoHandler(ctx *zero.Ctx) {
	var str string
	id := ctx.Event.UserID
	// UID
	userUid := genshin_cookie.GetUserUid(id)
	if len(userUid) >= 5 {
		str += "原神UID: " + userUid
	} else {
		str += "尚未绑定有效UID"
	}
	// Cookie
	userCookie := genshin_cookie.GetUserCookie(id)
	if len(userCookie) >= 10 {
		str += "\n已绑定米游社cookie"
	} else {
		str += "\n尚未绑定有效cookie"
	}
	// 自动签到
	eventFrom, err := GetEventFrom(ctx.Event.UserID)
	if err != nil {
		log.Warnf("GetEventFrom err: %v", err)
		str += "\n尚未开启米游社自动签到"
	} else if eventFrom.Auto {
		str += "\n已开启米游社自动签到"
	} else {
		str += "\n尚未开启米游社自动签到"
	}
	ctx.Send(str)
	return
}
