package bilibili

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "b站订阅",
	Usage: `订阅B站番剧、up主动态、直播，自动推送
用法：
	b站订阅番剧 [番剧名称或ID]: 订阅指定番剧或影视，支持按名称模糊搜索
	b站订阅up [up主名称或ID]：订阅指定up主的动态，支持用户名称模糊搜索
	b站订阅直播 [直播间ID]：订阅指定直播间的直播

	b站已有订阅：群聊中，展示该群所有群订阅；私聊中，展示你的所有个人订阅
	b站取消订阅 [订阅ID]：取消指定订阅，订阅ID请参照"b站已有订阅"中的订阅ID！

在私聊中调用时，代表个人订阅，只会私聊推送给你一个人
在群聊中调用时，代表群订阅（即会在该群中推送），需要拥有管理员权限`,
	SuperUsage: `
	b站全部订阅：（仅限私聊）展示所有用户、所有群的订阅
	b站取消订阅 [订阅ID] [QQ号]：取消指定用户的指定订阅；若QQ号为0，则取消该订阅ID下的所有订阅
	b站取消订阅 [订阅ID] 群[群号]：取消指定群的指定订阅
	b站cookie [你的b站cookie]：（仅限私聊）设置一个全局Cookie，全部cookie或仅SESSDATA皆可
即使不设置全局cookie，上述所有功能也可以正常使用，但设置后可以减小被b站限流的可能性；获取方法请自行百度
config-plugin配置项：
	bilibili.maxsearch: 最大搜索结果条数
	bilibili.group: 搜索结果以多少条为一组进行发送
	bilibili.limit: 动态更新推送时，动态内容最多展示多少字
	bilibili.picture: 动态更新推送时，图片动态最多发送多少张图片
	bilibili.link: 订阅内容更新时是(true)否(false)在消息中附加链接
	bilibili.atall: 群订阅内容更新时是(true)否(false)@全体成员`,
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
	proxy.OnFullMatch([]string{"b站全部订阅"}, zero.SuperUserPermission, zero.OnlyPrivate).
		SetBlock(true).SetPriority(3).Handle(allSubscribeHandler)
	proxy.OnCommands([]string{"b站cookie"}, zero.SuperUserPermission, zero.OnlyPrivate).
		SetBlock(true).SetPriority(3).Handle(cookieHandler)
	proxy.AddConfig("maxsearch", 8)
	proxy.AddConfig("group", 4)
	proxy.AddConfig("link", true)
	proxy.AddConfig("atAll", false)
	proxy.AddConfig("limit", 80)
	proxy.AddConfig("picture", 0)
	SetAPIDefault("search.type", "https://api.bilibili.com/x/web-interface/search/type")
	SetAPIDefault("bangumi.mdid", "https://api.bilibili.com/pgc/review/user")
	SetAPIDefault("user.info", "https://api.bilibili.com/x/space/acc/info")
	SetAPIDefault("user.dynamic", "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history")
	SetAPIDefault("live.info", "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom")
	// 初始化
	if cookie, err := proxy.GetLevelDB().Get([]byte("bilibili.cookie.global"), nil); err == nil && len(cookie) > 0 {
		SetGlobalCookie(string(cookie))
	}
	if len(AllSubscription()) > 0 {
		startPolling()
	}
}

func getContentLimit() int {
	l := proxy.GetConfigInt64("limit")
	if l <= 0 {
		return 0
	}
	return int(l)
}

var subscribeDealerMap = map[string]func(ctx *zero.Ctx, arg string, userID string){
	"番剧|动漫":  subscribeBangumi,
	"up主|up": subscribeUp,
	"直播间|直播": subscribeLive,
}

// 订阅处理
func subscribeHandler(ctx *zero.Ctx) {
	// 检查参数
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if args == "" { // 没有参数，多半是想查看已有订阅
		listSubscribeHandler(ctx)
		return
	}
	// 检查权限
	userID := strconv.FormatInt(ctx.Event.UserID, 10)
	if utils.IsMessageGroup(ctx) {
		if !auth.CheckPriority(ctx, 5, true) { // 群订阅权限检查
			ctx.Send("可以在私聊中开启个人订阅哦")
			return
		}
		userID = fmt.Sprintf("%v:%v", ctx.Event.GroupID, userID) // 群订阅：群号:发起用户ID
	}
	// 处理新订阅
	for k, dealer := range subscribeDealerMap {
		tps := strings.Split(k, "|")
		for _, tp := range tps {
			if strings.HasPrefix(args, tp) { // 处理特定类型订阅
				dealer(ctx, strings.TrimSpace(strings.TrimPrefix(args, tp)), userID)
				return
			}
		}
	}
	ctx.Send("只支持订阅番剧、up主、直播哦")
}

// 查看所有订阅处理
func allSubscribeHandler(ctx *zero.Ctx) {
	subs := AllSubscription()
	for _, sub := range subs {
		msg := sub.GenMessage(true)
		if len(msg) > 0 {
			ctx.Send(msg)
		}
	}
	if len(subs) == 0 {
		ctx.Send("暂时没有b站订阅")
	}
}

// 查看已有订阅处理
func listSubscribeHandler(ctx *zero.Ctx) {
	var subs []Subscription
	if utils.IsMessageGroup(ctx) {
		subs = GetSubForGroup(ctx.Event.GroupID)
	} else {
		subs = GetSubForPrimary(ctx.Event.UserID)
	}
	for _, sub := range subs {
		msg := sub.GenMessage(false)
		if len(msg) > 0 {
			ctx.Send(msg)
		}
	}
	if len(subs) == 0 {
		ctx.Send("暂时没有b站订阅")
	}
}

// 取消订阅处理
func unsubscribeHandler(ctx *zero.Ctx) {
	// 群取消订阅权限检查
	if utils.IsMessageGroup(ctx) && !auth.CheckPriority(ctx, 5, true) {
		return
	}
	// 参数检查
	args := strings.Split(strings.TrimSpace(utils.GetArgs(ctx)), " ")
	if len(args) == 0 {
		ctx.Send("参数不对哦，可以看看帮助")
		return
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		log.Warnf("wrong id, err : %v", err)
		ctx.Send("订阅ID格式不对哦，可以看看帮助")
		return
	}
	// 处理
	if len(args) >= 2 && utils.IsSuperUser(ctx.Event.UserID) {
		if strings.HasPrefix(args[1], "群") { // 超级用户指定群
			group, err := strconv.ParseInt(args[1][len("群"):], 10, 64)
			if err != nil {
				log.Warnf("wrong group id, err : %v", err)
				ctx.Send("群号格式不对哦，可以看看帮助")
				return
			}
			err = DeleteSubscription(Subscription{ID: int(id), SubUsers: strconv.FormatInt(group, 10) + ":"})
			if err != nil {
				log.Errorf("DeleteSubscription err: %v", err)
				ctx.Send("失败了...")
				return
			}
		} else { // 超级用户指定QQ
			user, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Warnf("wrong user id, err : %v", err)
				ctx.Send("QQ号格式不对哦，可以看看帮助")
				return
			}
			userID := strconv.FormatInt(user, 10)
			if user == 0 {
				userID = SubUserAll
			}
			err = DeleteSubscription(Subscription{ID: int(id), SubUsers: userID})
			if err != nil {
				log.Errorf("DeleteSubscription err: %v", err)
				ctx.Send("失败了...")
				return
			}
		}
		ctx.Send("好哒")
		return
	}
	// 普通用户
	userID := strconv.FormatInt(ctx.Event.UserID, 10)
	if utils.IsMessageGroup(ctx) {
		userID = strconv.FormatInt(ctx.Event.GroupID, 10) + ":"
	}
	err = DeleteSubscription(Subscription{ID: int(id), SubUsers: userID})
	if err != nil {
		log.Errorf("DeleteSubscription err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("好哒")
}

// 设置全局cookie处理
func cookieHandler(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if len(arg) > 0 { // 测试cookie可用性
		c := NewClient()
		c.SetCookie(regularBilibiliCookie(arg))
		res, err := c.GetGJson("https://api.bilibili.com/x/space/upstat?mid=456664753")
		if err != nil {
			log.Errorf("check error: %v", err)
			ctx.Send("失败了...")
			return
		}
		if res.Get("code").Int() != 0 || !res.Get("data.likes").Exists() {
			log.Errorf("check response code=%d, message=%s", res.Get("code").Int(), res.Get("message").String())
			ctx.Send("此cookie无效")
			return
		}
	}
	// 设置
	if err := proxy.GetLevelDB().Put([]byte("bilibili.cookie.global"), []byte(arg), nil); err != nil {
		log.Errorf("leveldb save error: %v", err)
		ctx.Send("失败了...")
		return
	}
	SetGlobalCookie(arg)
	ctx.Send("好哒")
}

// 订阅番剧处理
func subscribeBangumi(ctx *zero.Ctx, arg string, userID string) {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		// 搜索相关番剧
		s, err := NewSearch().Bangumi(arg)
		if err != nil {
			log.Errorf("bilibili search bangumi error: %v", err)
			ctx.Send("失败了...")
			return
		}
		// 同时尝试搜索影视
		fts, err := NewSearch().FT(arg)
		if err != nil { // 不返回
			log.Errorf("bilibili search FT error: %v", err)
		}
		s = append(s, fts...)
		if len(s) == 0 {
			ctx.Send("没有找到相关的番剧")
			return
		}
		id = s[0].MediaID
		// 选择番剧
		if len(s) > 1 {
			maxSearch := int(proxy.GetConfigInt64("maxsearch"))
			if maxSearch > 0 && len(s) > maxSearch { // 限定最大结果条数
				s = s[:maxSearch]
			}
			// 发送搜索结果
			ms := make([]messager, len(s))
			for i := range s {
				ms[i] = s[i]
			}
			SendSearchResult(ctx, ms)
			ctx.Send("如果上述番剧中有你想订阅的番剧，请答复其序号（方括号内）；若没有，请说没有")
			event := utils.WaitNextMessage(ctx)
			if event == nil {
				ctx.Send("那算啦")
				return
			}
			index, err := strconv.Atoi(strings.TrimSpace(event.Message.ExtractPlainText()))
			if err != nil {
				ctx.Send("那算啦")
				return
			}
			if index <= 0 || index > len(s) {
				ctx.Send("没有这个序号的番剧")
				return
			}
			id = s[index-1].MediaID
		}
	}
	// 获取番剧信息
	i, err := NewBangumi().ByMDID(id)
	if err != nil {
		log.Errorf("bilibili get bangumi info by MDID error: %v", err)
		ctx.Send("失败了...")
		return
	}
	// 确定订阅
	if isConfirm(ctx, fmt.Sprintf("是否订阅番剧：%v", i.Title)) {
		err := AddSubscription(Subscription{
			SubType:          SubTypeBangumi,
			SubUsers:         userID,
			BID:              id,
			BangumiLastIndex: i.NewEP.Name,
		})
		if err != nil {
			log.Errorf("AddSubscription err: %v", err)
			ctx.Send("失败了...")
			return
		}
		ctx.Send("好哒")
		return
	}
	ctx.Send("那算啦")
}

// 订阅up主动态处理
func subscribeUp(ctx *zero.Ctx, arg string, userID string) {
	// 解析参数
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		// 搜索相关UP主
		us, err := NewSearch().User(arg)
		if err != nil {
			log.Errorf("bilibili search user error: %v", err)
			ctx.Send("失败了...")
			return
		}
		if len(us) == 0 {
			ctx.Send("没有找到相关的UP主")
			return
		}
		id = us[0].MID
		// 选择UP主
		if len(us) > 1 {
			maxSearch := int(proxy.GetConfigInt64("maxsearch"))
			if maxSearch > 0 && len(us) > maxSearch { // 限定最大结果条数
				us = us[:maxSearch]
			}
			// 发送搜索结果
			ms := make([]messager, len(us))
			for i := range us {
				ms[i] = us[i]
			}
			SendSearchResult(ctx, ms)
			ctx.Send("如果上述UP主中有你想订阅的UP主，请答复其序号（方括号内）；若没有，请说没有")
			event := utils.WaitNextMessage(ctx)
			if event == nil {
				ctx.Send("那算啦")
				return
			}
			index, err := strconv.Atoi(strings.TrimSpace(event.Message.ExtractPlainText()))
			if err != nil {
				ctx.Send("那算啦")
				return
			}
			if index <= 0 || index > len(us) {
				ctx.Send("没有这个序号的UP主")
				return
			}
			id = us[index-1].MID
		}
	}
	// 获取UP主信息
	i, err := NewUser(id).Info()
	if err != nil {
		log.Errorf("bilibili get user info by ID error: %v", err)
		ctx.Send("失败了...")
		return
	}
	// 确定订阅
	subStr := fmt.Sprintf("id=%v", i.MID)
	if i.LiveRoomID > 0 {
		subStr += fmt.Sprintf(",直播间ID=%v", i.LiveRoomID)
	}
	if isConfirm(ctx, fmt.Sprintf("是否订阅UP主：%v(%s)", i.Name, subStr)) {
		err := AddSubscription(Subscription{
			SubType:  SubTypeUp,
			SubUsers: userID,
			BID:      id,
		})
		if err != nil {
			log.Errorf("AddSubscription err: %v", err)
			ctx.Send("失败了...")
			return
		}
		ctx.Send("好哒")
		return
	}
	ctx.Send("那算啦")
}

// 订阅直播处理
func subscribeLive(ctx *zero.Ctx, arg string, userID string) {
	// 解析参数
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		log.Errorf("直播间ID参数错误：%v", err)
		ctx.Send("直播间ID格式不对哦，可以看看帮助")
		return
	}
	// 获取直播间信息
	l, err := NewLiveRoom(id).Info()
	if err != nil {
		log.Errorf("bilibili get live room info by ID error: %v", err)
		ctx.Send("失败了...")
		return
	}
	// 确定订阅
	if isConfirm(ctx, fmt.Sprintf("是否订阅%v的直播间(%v)", l.Anchor.Name, id)) {
		err := AddSubscription(Subscription{
			SubType:    SubTypeLive,
			SubUsers:   userID,
			BID:        id,
			LiveStatus: l.IsOpen(),
		})
		if err != nil {
			log.Errorf("AddSubscription err: %v", err)
			ctx.Send("失败了...")
			return
		}
		ctx.Send("好哒")
		return
	}
	ctx.Send("那算啦")
}

func isConfirm(ctx *zero.Ctx, tip string) bool {
	ctx.Send(tip)
	event := utils.WaitNextMessage(ctx)
	if event == nil {
		return false
	}
	confirm := strings.TrimSpace(event.Message.ExtractPlainText())
	if confirm == "是" || confirm == "确定" || confirm == "确认" {
		return true
	}
	return false
}
