package emoji_mix

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "混合表情",
	Usage: `
混合两个表情, 表情最好是emoji表情, qq表情没有完全适配
用法：
	[emoji表情1][emoji表情2]
	
`,
	Classify: "一般功能",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnMessage(match).SetBlock(true).SetPriority(4).Handle(handleMixMsg)
}

func handleMixMsg(ctx *zero.Ctx) {
	if len(ctx.Event.Message) == 2 {
		getMixAndSend(ctx, ctx.Event.Message)
	} else {
		return
	}
}

func getMixAndSend(ctx *zero.Ctx, msg message.Message) {
	var r1 = face2emoji(msg[0])
	var r2 = face2emoji(msg[1])
	log.Debugln("[emojimix] match:", msg)
	u1 := fmt.Sprintf(bed, emojis[r1], r1, r1, r2)
	u2 := fmt.Sprintf(bed, emojis[r2], r2, r2, r1)
	log.Debugln("[emojimix] u1:", u1)
	log.Debugln("[emojimix] u2:", u2)
	resp1, err := http.Head(u1)
	if err == nil {
		err := resp1.Body.Close()
		if err != nil {
			log.Warnf("<emojimix> Body 关闭错误")
		}
		if resp1.StatusCode == http.StatusOK {
			ctx.SendChain(message.Image(u1))
			return
		}
	}
	resp2, err := http.Head(u2)
	if err == nil {
		err := resp2.Body.Close()
		if err != nil {
			log.Warnf("<emojimix> Body 关闭错误")
		}
		if resp2.StatusCode == http.StatusOK {
			ctx.SendChain(message.Image(u2))
			return
		}
	}
}

const bed = "https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png"

//copy from zeroBot
func match(ctx *zero.Ctx) bool {
	log.Debugln("[emojimix] msg:", ctx.Event.Message)
	if len(ctx.Event.Message) == 2 {
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
	log.Debugln("[emojimix] raw msg:", ctx.Event.RawMessage)
	if len(r) == 2 {
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

func face2emoji(face message.MessageSegment) rune {
	if face.Type == "face" {
		id, err := strconv.Atoi(face.Data["id"])
		if err != nil {
			return 0
		}
		if r, ok := qqface[id]; ok {
			return r
		}
	}
	if face.Type == "text" {
		var id, _ = utf8.DecodeRuneInString(strings.Split(face.Data["text"], "")[1])
		if r, ok := qqface[int(id)]; ok {
			return r
		}
	}

	return 0
}