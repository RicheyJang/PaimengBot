package poke

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/ban"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/pixiv"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:  "戳一戳",
	Usage: `戳一戳\拍一拍时随机回复`,
	SuperUsage: `需要把go-cqhttp的设备类型修改为非手表
config-plugin配置项：
	poke.replies: 回复内容列表，会从中随机选取，支持CQ码
另外，回复内容里还可以附带上某种动作，例如：
不准戳我！[mute] ：这将回复一句"不准戳我"，并将其禁言2分钟
目前支持的动作：
	[mute]：禁言两分钟，前提是将Bot设为群管理员
	[poke]：戳回去
	[ban]：封禁两分钟
	[pixiv]：随机一张好康的图片
另另外，动作内还可以附带一个0~1的实数，代表该动作的发生概率，例如：
不准戳我！[mute 0.5] ：当随机选取到这一回复时，会有50%的概率将其禁言2分钟`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnNotice(func(ctx *zero.Ctx) bool {
		return ctx.Event.NoticeType == "notify" && ctx.Event.SubType == "poke" && ctx.Event.TargetID == ctx.Event.SelfID
	}).SetBlock(true).ThirdPriority().Handle(pokeHandler)
	proxy.AddConfig(consts.PluginConfigCDKey, "3s")
	proxy.AddConfig("replies", []string{"？", "hentai!", "( >﹏<。)", "好气喔，我要给你起个难听的绰号", "那...那里...那里不能戳...", "[pixiv]喏", "不准戳我啦[mute 0.1]", "[poke]"})
}

var actionRegex = regexp.MustCompile(`\[([a-z]{2,10})\s*([01]\.\d+)?]`)

func pokeHandler(ctx *zero.Ctx) {
	// 初始化
	if proxy.LockUser(ctx.Event.UserID) {
		return
	}
	defer proxy.UnlockUser(ctx.Event.UserID)
	replies := proxy.GetConfigStrings("replies")
	if len(replies) == 0 {
		return
	}
	reply := replies[rand.Intn(len(replies))]
	log.Info("即将回复：", reply)
	// 发送回复内容
	str := strings.TrimSpace(actionRegex.ReplaceAllString(reply, ""))
	index := 0
	if loc := actionRegex.FindStringIndex(reply); len(loc) > 0 {
		index = loc[0]
	}
	if len(str) > 0 && index > 0 { // 内容非空 且 先回复再动作
		ctx.Send(str)
	}
	// 解析、执行动作字符串
	actions := actionRegex.FindAllStringSubmatch(reply, -1)
	for _, action := range actions {
		if len(action) <= 2 {
			continue
		}
		rate, _ := strconv.ParseFloat(action[2], 32)
		dealActions(ctx, action[1], rate)
	}
	if len(str) > 0 && index == 0 { // 内容非空 且 （先动作再回复 或 无动作）
		ctx.Send(str)
	}
}

// 处理动作
func dealActions(ctx *zero.Ctx, action string, rate float64) {
	if rate != 0 && rand.Float64() > rate { // 不满足概率，不触发
		return
	}
	if ctx.Event.UserID == 0 {
		return
	}
	log.Infof("do %v to %v", action, ctx.Event.UserID)
	// 执行各类动作
	switch action {
	case "mute": // 禁言
		if ctx.Event.GroupID == 0 || utils.IsSuperUser(ctx.Event.UserID) {
			break
		}
		ctx.SetGroupBan(ctx.Event.GroupID, ctx.Event.UserID, 120)
	case "poke": // 戳一戳
		if ctx.Event.GroupID == 0 {
			break
		}
		ctx.Send(message.Poke(ctx.Event.UserID))
	case "ban": // 封禁
		if utils.IsSuperUser(ctx.Event.UserID) {
			break
		}
		_ = ban.SetUserPluginStatus(false, ctx.Event.UserID, nil, 2*time.Minute)
	case "pixiv": // 好康的
		pics := pixiv.GetRandomPictures(nil, 1, false)
		for _, pic := range pics {
			if msg, err := pic.GenSinglePicMsg(); err == nil {
				ctx.Send(msg)
				break
			}
		}
	}
}
