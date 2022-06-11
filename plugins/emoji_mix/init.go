package emoji_mix

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/chat"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "混合表情",
	Usage: `
	混合两个表情，表情最好是emoji表情，qq表情没有完全适配
用法：
	[emoji表情1][emoji表情2]: 合成两个emoji表情`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(match).SetBlock(true).SetPriority(4).Handle(handleMixMsg)
}

func handleMixMsg(ctx *zero.Ctx) {
	// 初始化
	r, ok := ctx.State["emojimix"].([]rune)
	if !ok || len(r) != 2 { // something wrong 防止panic
		return
	}
	var r1 = r[0]
	var r2 = r[1]
	// 分别尝试两种合成
	u1 := fmt.Sprintf(bedURL, emojis[r1], r1, r1, r2)
	u2 := fmt.Sprintf(bedURL, emojis[r2], r2, r2, r1)
	resp1, err := http.Head(u1)
	if err == nil { // 第一种成功
		defer resp1.Body.Close()
		if resp1.StatusCode == http.StatusOK {
			ctx.SendChain(message.Image(u1))
			return
		}
	}
	resp2, err := http.Head(u2)
	if err == nil { // 第二种成功
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusOK {
			ctx.SendChain(message.Image(u2))
			return
		}
	}
	chat.IDoNotKnow(ctx, string(r))
}

const bedURL = "https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png"

// match 判断是否为可以混合的表情 并将emoji索引保存到State
func match(ctx *zero.Ctx) bool {
	if len(ctx.Event.Message) == 2 { //两个qq表情或者qq emoji混合表情
		r1 := face2emoji(ctx.Event.Message[0])
		if _, ok := emojis[r1]; !ok {
			return false
		}
		r2 := face2emoji(ctx.Event.Message[1])
		if _, ok := emojis[r2]; !ok {
			return false
		}
		ctx.State["emojimix"] = []rune{r1, r2}
		return true
	}

	r := []rune(ctx.Event.RawMessage)
	if len(r) == 2 { //纯emoji
		if _, ok := emojis[r[0]]; !ok {
			return false
		}
		if _, ok := emojis[r[1]]; !ok {
			return false
		}
		ctx.State["emojimix"] = r
		return true
	}
	return false
}

// 获取qq表情 或 emoji对应的索引
func face2emoji(face message.MessageSegment) rune {
	if face.Type == "face" {
		id, err := strconv.Atoi(face.Data["id"])
		if err != nil {
			return 0
		}
		return qqface[id]
	} else if face.Type == "text" {
		rs := []rune(face.Data["text"])
		//防止qq表情和一堆emoji也被识别
		if len(rs) != 1 {
			return 0
		}
		return rs[0]
	}
	return 0
}
