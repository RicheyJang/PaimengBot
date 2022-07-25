package whatpicture

import (
	"time"

	"github.com/RicheyJang/PaimengBot/basic/sc"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "搜图",
	Usage: `用法：
	搜图/搜本子 [图片]: 搜索该图出现的插画、本子等信息`,
	SuperUsage: `请预先配置saucenao的API Key，否则无法使用
config-plugin配置项：
	whatpicture.max: 每个用户的每日最大搜索次数
	whatpicture.timeout: 单次搜索超时时间
	whatpicture.link: 搜索结果中是(true)否(false)附带链接
	whatpicture.key: saucenao的API Key
saucenao的API Key可以在https://saucenao.com/user.php?page=search-api中注册获取`,
	Classify: "好康的",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"搜图", "识图", "搜本子"}).SetBlock(true).SetPriority(4).Handle(searchPicHandler)
	proxy.AddConfig("max", 3)
	proxy.AddConfig("key", "")
	proxy.AddConfig("timeout", "30s")
	proxy.AddConfig("link", true)
	proxy.AddConfig(consts.PluginConfigCDKey, "1m")
	proxy.SetCallLimiter("all", 24*time.Hour, 5).BindTimesConfig("max").SkipSuperuser(true)
}

func searchPicHandler(ctx *zero.Ctx) {
	// 检测API Key是否配置
	if len(proxy.GetConfigString("key")) == 0 {
		if utils.IsSuperUser(ctx.Event.UserID) {
			ctx.Send(`请配置saucenao的API Key，可以看看"帮助 搜图"`)
		} else {
			ctx.Send("管理员配置相关的API Key后才可以搜图哦")
		}
		sc.SetNeedReturnCost(ctx)
		return
	}
	// 获取图片
	urls := utils.GetImageURLs(ctx.Event)
	if len(urls) == 0 { // 没有发图，等待他的下一条消息
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("图呢？"))
		urls = utils.GetImageURLs(utils.WaitNextMessage(ctx))
		if len(urls) == 0 { // 依旧没有发图
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("那就算啦"))
			return
		}
	}
	// 上锁，防止重复调用
	if proxy.LockUser(0) {
		ctx.Send("有请求正在处理中哦")
		sc.SetNeedReturnCost(ctx)
		return
	}
	defer proxy.UnlockUser(0)
	// 检测是否到限额
	if !proxy.CheckCallLimit("all", ctx.Event.UserID) {
		ctx.Send("今日搜图次数已达上限，请明天再试")
		sc.SetNeedReturnCost(ctx)
		return
	}
	// 只查询第一张图
	msgs, err := SearchPicture(urls[0], utils.IsMessagePrimary(ctx))
	if err != nil {
		log.Warnf("SearchPicture err: user=%v,pic=%v,err=%v", ctx.Event.UserID, urls[0], err)
		ctx.Send(append(message.Message{message.At(ctx.Event.UserID)}, msgs[0]...))
		return
	}
	if utils.IsMessageGroup(ctx) { // 群消息 节点转发
		var forwardMsg message.Message
		for _, m := range msgs {
			forwardMsg = append(forwardMsg, message.CustomNode(utils.GetBotNickname(), ctx.Event.SelfID, m))
		}
		ctx.SendGroupForwardMessage(ctx.Event.GroupID, forwardMsg)
	} else {
		for _, m := range msgs {
			ctx.Send(m)
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// SearchPicture 搜图，参数：url为图片链接，返回整理后需要发出的消息体
func SearchPicture(url string, showAdult bool) ([]message.Message, error) {
	return searchPicBySaucenao(url, showAdult)
}
