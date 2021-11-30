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

func banUser(ctx *zero.Ctx) {
	dealBanUser(false, ctx)
}

func unbanUser(ctx *zero.Ctx) {
	dealBanUser(true, ctx)
}

func dealBanArgs(ctx *zero.Ctx) (userID int64, plugin *manager.PluginCondition, period time.Duration, err error) {
	preArgs := utils.GetArgs(ctx)
	// 检查
	args := strings.Split(strings.TrimSpace(preArgs), " ")
	if len(args) == 0 || len(args[0]) == 0 {
		ctx.Send("你倒是告诉我封掉谁呀")
		return 0, nil, 0, fmt.Errorf("no such user")
	}
	// 处理用户ID
	userID, err = strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		ctx.Send("用户ID格式不对哦")
		return 0, nil, 0, fmt.Errorf("no such user %v", args[0])
	}
	// 处理时长
	if len(args) >= 2 {
		period, err = time.ParseDuration(args[len(args)-1])
		if err != nil {
			period = 0
		} else {
			args = args[:len(args)-1]
		}
	}
	// 处理插件名
	if len(args) >= 2 {
		plugin = findPluginByName(args[len(args)-1])
		if plugin == nil {
			ctx.Send(fmt.Sprintf("没有叫%v的功能哦，可以看看帮助", args[len(args)-1]))
			return 0, nil, period, fmt.Errorf("no such plugin %v", args[len(args)-1])
		}
	}
	log.Debugf("dealBanArgs res: userID=%v,plugin=%v,period=%v,err=%v", userID, plugin, period, err)
	return
}

func dealBanUser(status bool, ctx *zero.Ctx) {
	if utils.IsMessagePrimary(ctx) {
		if !utils.IsSuperUser(ctx.Event.UserID) {
			ctx.Send("请在群聊中封禁/解封功能哦，或者联系管理员")
			return
		}
		// 个人开关
		userID, plugin, period, err := dealBanArgs(ctx)
		if err != nil {
			log.Errorf("dealBanUser err: %v", err)
			return
		}
		dealUserPluginStatus(ctx, status, userID, plugin, period)
		return
	}
	userID, plugin, period, err := dealBanArgs(ctx)
	if err != nil {
		log.Errorf("dealBanUser err: %v", err)
		return
	}
	// 检查是否为本群成员
	res := ctx.GetGroupMemberInfo(ctx.Event.GroupID, userID, true)
	if res.Get("group_id").Int() != ctx.Event.GroupID {
		ctx.Send("这谁啊？")
		return
	}
	if !utils.IsSuperUser(ctx.Event.UserID) && utils.IsSuperUser(userID) {
		ctx.Send("？")
		return
	}
	// 更新封禁状态
	dealUserPluginStatus(ctx, status, userID, plugin, period)
}
