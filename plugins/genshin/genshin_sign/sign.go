package genshin_sign

import (
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "米游社签到",
	Usage: `需要预先绑定cookie和uid，参见：帮助 米游社管理
用法：
	米游社签到：顾 名 思 义
	米游社定时签到 [打开/关闭]：即可打开/关闭米游社自动定时签到，仅限好友私聊
	是否已开启米游社定时签到：看看你有没有打开自动定时签到，仅限好友私聊`,
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
	proxy.OnFullMatch([]string{"米游社签到"}).SetBlock(true).SetPriority(3).Handle(singleSignHandler)
	// 防止群消息乱飞，目前仅允许私聊使用自动签到
	proxy.OnCommands([]string{"米游社定时签到", "米游社自动签到"}, zero.OnlyPrivate).SetBlock(true).SetPriority(3).Handle(autoSignHandler)
	proxy.OnFullMatch([]string{"是否已开启米游社定时签到", "是否已开启米游社自动签到"}, zero.OnlyPrivate).SetBlock(true).SetPriority(3).Handle(queryAutoHandler)
	proxy.AddConfig("daily.hour", 9)
	proxy.AddConfig("daily.min", 0)
	// 添加定时签到任务
	taskID, _ = proxy.AddScheduleDailyFunc(
		int(proxy.GetConfigInt64("daily.hour")),
		int(proxy.GetConfigInt64("daily.min")),
		autoSignTask)
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
		ctx.Send(images.GenStringMsg(msg))
	}
	ctx.Send(message.Text(msg))
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
	eventFrom, err := GetEventFrom(ctx.Event.UserID)
	if err != nil {
		log.Warnf("GetEventFrom err: %v", err)
		ctx.Send("尚未开启")
		return
	}
	if eventFrom.Auto {
		ctx.Send("已开启")
	} else {
		ctx.Send("尚未开启")
	}
	return
}
