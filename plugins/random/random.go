package random

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/nickname"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "随机",
	Usage: `交给命运吧！
用法：
	随机 [选项]+：从多个选项中随机答复一个，选项请以空格分隔
	随机数 [范围]：从指定范围中随机答复一个数字，范围格式参见示例
示例：
	随机数 1...10：从1到10中随机答复一个数字`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"随机数"}).SetBlock(true).SetPriority(3).Handle(randomNumHandler)
	proxy.OnCommands([]string{"随机"}).SetBlock(true).SetPriority(4).Handle(randomItemHandler)
}

func randomNumHandler(ctx *zero.Ctx) {
	args := strings.TrimSpace(utils.GetArgs(ctx))
	if len(args) == 0 {
		ctx.Send("？")
		return
	}

	min, max := 0, 0
	_, err := fmt.Sscanf(args, "%d...%d", &min, &max)
	if err != nil {
		log.Warnf("Sscanf error: %v", err)
		ctx.Send("参数不对哦")
		return
	}
	if min > max {
		min, max = max, min
	}

	nick := nickname.GetNickname(ctx.Event.UserID, "你")
	tip := fmt.Sprintf("命立天地宇宙，芬芳乐园伊甸，让大幻梦森罗万象狂气断罪眼指引迷途之人吧！%v的命运乃：", nick)
	tip += strconv.Itoa(min + rand.Intn(max-min+1))
	if utils.IsMessageGroup(ctx) {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(tip))
	} else {
		ctx.Send(tip)
	}
}

func randomItemHandler(ctx *zero.Ctx) {
	args := strings.TrimSpace(utils.GetArgs(ctx))
	items := utils.MergeStringSlices(strings.Split(args, " "))
	if len(items) == 0 {
		ctx.Send("？")
		return
	}

	nick := nickname.GetNickname(ctx.Event.UserID, "你")
	tip := fmt.Sprintf("命立天地宇宙，芬芳乐园伊甸，让大幻梦森罗万象狂气断罪眼指引迷途之人吧！%v的命运乃：", nick)
	tip += items[rand.Intn(len(items))]
	if utils.IsMessageGroup(ctx) {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(tip))
	} else {
		ctx.Send(tip)
	}
}
