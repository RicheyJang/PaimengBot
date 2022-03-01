package bilibili

import (
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "b站订阅",
	Usage: `订阅B站番剧、up主动态、直播，自动推送
用法：
	b站订阅番剧 [番剧名称或ID]: 订阅指定番剧，支持番剧名称模糊搜索
	b站订阅up [up主ID]：订阅指定up主的动态
	b站订阅直播 [直播间ID]：订阅指定直播间的直播

	b站已有订阅：展示所有订阅
	b站取消订阅 [订阅ID]：取消指定订阅，订阅ID请参照"b站已有订阅"

在私聊中调用时，代表个人订阅，只会私聊推送给你一个人
在群聊中调用时，代表群订阅（即会在该群中推送），需要拥有管理员权限`,
	Classify: "实用工具",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"b站订阅"}).SetBlock(true).SetPriority(3).Handle(subscribeHandler)
	proxy.OnFullMatch([]string{"b站已有订阅"}).SetBlock(true).SetPriority(3).Handle(listSubscribeHandler)
	proxy.OnCommands([]string{"b站取消订阅"}).SetBlock(true).SetPriority(3).Handle(unsubscribeHandler)
	SetAPIDefault("search.type", "https://api.bilibili.com/x/web-interface/search/type")
	SetAPIDefault("bangumi.mdid", "https://api.bilibili.com/pgc/review/user")
	SetAPIDefault("user.info", "https://api.bilibili.com/x/space/acc/info")
	SetAPIDefault("user.dynamic", "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history")
	SetAPIDefault("live.info", "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom")
}

var subscribeDealerMap = map[string]func(*zero.Ctx, string){
	"番剧|动漫":  subscribeBangumi,
	"up主|up": subscribeUp,
	"直播间|直播": subscribeLive,
}

// 订阅处理
func subscribeHandler(ctx *zero.Ctx) {
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if args == "" { // 没有参数，多半是想查看已有订阅
		listSubscribeHandler(ctx)
		return
	}
	// TODO 群订阅权限检查
	for k, dealer := range subscribeDealerMap {
		tps := strings.Split(k, "|")
		for _, tp := range tps {
			if strings.HasPrefix(args, tp) { // 处理特定类型订阅
				dealer(ctx, strings.TrimSpace(strings.TrimPrefix(args, tp)))
				return
			}
		}
	}
	ctx.Send("只支持订阅番剧、up主、直播哦")
}

// 查看已有订阅处理
func listSubscribeHandler(ctx *zero.Ctx) {
	// TODO implement me
}

// 取消订阅处理
func unsubscribeHandler(ctx *zero.Ctx) {
	// TODO implement me
}

// 订阅番剧处理
func subscribeBangumi(ctx *zero.Ctx, arg string) {
	// TODO implement me
}

// 订阅up主动态处理
func subscribeUp(ctx *zero.Ctx, arg string) {
	// TODO implement me
}

// 订阅直播处理
func subscribeLive(ctx *zero.Ctx, arg string) {
	// TODO implement me
}
