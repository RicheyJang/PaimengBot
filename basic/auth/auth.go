package auth

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "权限鉴权",
	Usage: `
用法：
	更新管理员权限：会将所有群中未被设置权限的管理员设为默认权限
	设置管理员权限 [群号] [用户ID] [Level]：将指定群的指定用户权限设为Level
备注：
	每日1点5分，会更新所有群管理员权限; 
	群管理员变动时会自动刷新该群权限，并清除被撤下的管理员的所有权限;（event包）
	权限level(>=1)数字越小，权限越高
	权限level设为0代表清除该用户权限，该用户无管理员权限
`,
	IsSuperOnly: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"更新管理员权限"}).SetBlock(true).FirstPriority().Handle(flushAllPriority)
	proxy.OnCommands([]string{"设置管理员权限"}).SetBlock(true).FirstPriority().Handle(setOnePriority)
	proxy.AddConfig("defaultLevel", 5)
	proxy.AddConfig("ownerLevel", 1) // 群主的默认权限等级
	proxy.AddConfig("superLevel", 1) // 超级用户的默认权限等级
	_, _ = proxy.AddScheduleDailyFunc(1, 5, initialAllPriority)
	manager.AddPreHook(authHook) // 在调用插件前检查管理员权限
}

// Hook 在调用插件前检查管理员权限
func authHook(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	if condition.AdminLevel == 0 { // 插件未设置权限
		return nil
	}
	if ctx.Event == nil || ctx.Event.UserID == 0 {
		return nil
	}
	if !utils.IsMessage(ctx) { // 非消息事件
		return nil
	}
	level := GetGroupUserPriority(ctx.Event.GroupID, ctx.Event.UserID)
	if level <= 0 || utils.IsGroupAnonymous(ctx) { // 无权限或匿名消息，权限设为最低
		level = math.MaxInt
	}
	if level > condition.AdminLevel {
		if level == math.MaxInt {
			ctx.Send(fmt.Sprintf("你的权限不足喔，需要权限%v", condition.AdminLevel))
		} else {
			ctx.Send(fmt.Sprintf("你的权限(%v)不足喔，需要权限%v", level, condition.AdminLevel))
		}
		return errors.New("用户权限不足")
	}
	return nil
}

func flushAllPriority(ctx *zero.Ctx) {
	initialAllPriority()
	ctx.Send("好哒")
}

func setOnePriority(ctx *zero.Ctx) {
	org := utils.GetArgs(ctx)
	args := strings.Split(strings.TrimSpace(org), " ")
	if len(args) < 3 {
		ctx.Send("参数不够哦，可以参考一下帮助")
		return
	}
	groupID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		ctx.Send("群号格式不对哦")
		return
	}
	userID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		ctx.Send("用户ID格式不对哦")
		return
	}
	level, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil || level < 0 {
		ctx.Send("权限等级格式不对哦")
		return
	}
	err = SetGroupUserPriority(groupID, userID, int(level))
	if err != nil {
		log.Errorf("SetGroupUserPriority err: %v", err)
		ctx.Send("设置失败了...")
	} else {
		ctx.Send(fmt.Sprintf("成功把%v在群%v的权限设置成%v啦", userID, groupID, level))
	}
}

// 每天定时初始化所有未被设置权限的管理员为默认权限
func initialAllPriority() {
	ctx := utils.GetBotCtx()
	if ctx == nil {
		log.Errorf("initialAllPriority err: zero.Ctx == nil")
		return
	}
	errCount := 0
	res := ctx.GetGroupList()
	groups := res.Array()
	for _, group := range groups {
		if err := InitialGroupPriority(ctx, group.Get("group_id").Int()); err != nil {
			log.Warnf("更新群(%v)管理员权限出错：%v", group.Get("group_id").Int(), err)
			errCount += 1
		}
	}
	log.Infof("更新全部群管理员权限完成，共%v个，失败%v个", len(groups), errCount)
}
