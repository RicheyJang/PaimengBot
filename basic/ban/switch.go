package ban

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func openPlugin(ctx *zero.Ctx) {
	switchPlugin(true, ctx)
}

func closePlugin(ctx *zero.Ctx) {
	switchPlugin(false, ctx)
}

func dealSwitchArgs(ctx *zero.Ctx, dealGroup bool) (
	groupID int64, plugin *manager.PluginCondition, period time.Duration, err error) {
	preArgs := utils.GetArgs(ctx)
	// 检查
	args := strings.Split(strings.TrimSpace(preArgs), " ")
	if len(args) == 0 || len(args[0]) == 0 {
		ctx.Send("你倒是告诉我开关什么功能呀")
		return 0, nil, 0, fmt.Errorf("no such plugin")
	}
	// 处理群号
	if dealGroup && len(args) >= 2 {
		groupID, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			groupID = 0
		} else {
			args = args[1:]
		}
	}
	// 处理时长
	if len(args) >= 2 {
		period, err = time.ParseDuration(args[len(args)-1])
		if err != nil {
			ctx.Send("时间格式不对哦，可以看看帮助")
			return
		}
	}
	// 处理插件名
	plugin = findPluginByName(args[0])
	if plugin == nil {
		ctx.Send(fmt.Sprintf("没有叫%v的功能哦，可以看看帮助", args[0]))
		return 0, nil, period, fmt.Errorf("no such plugin %v", args[0])
	}
	log.Debugf("dealSwitchArgs res: groupID=%v,plugin=%v,period=%v,err=%v", groupID, plugin, period, err)
	return
}

func switchPlugin(status bool, ctx *zero.Ctx) {
	if ctx.Event.MessageType == "private" {
		if !utils.IsSuperUser(ctx.Event.UserID) {
			ctx.Send("请在群聊中开关功能哦，或者联系管理员")
			return
		}
		// 全局开关
		groupID, plugin, period, err := dealSwitchArgs(ctx, true)
		if err != nil {
			log.Errorf("switchPlugin err: %v", err)
			return
		}
		if groupID > 0 {
			dealGroupPluginStatus(ctx, status, groupID, plugin, period)
		} else {
			dealUserPluginStatus(ctx, status, 0, plugin, period)
		}
		return
	}
	// 群聊
	_, plugin, period, err := dealSwitchArgs(ctx, false)
	if err != nil {
		log.Errorf("switchPlugin err: %v", err)
		return
	}
	dealGroupPluginStatus(ctx, status, ctx.Event.GroupID, plugin, period)
}
