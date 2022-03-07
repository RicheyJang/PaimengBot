package bilibili

import (
	"fmt"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type messager interface {
	GenText(index int) string
	GenMessage(index int) message.Message
}

func SendSearchResult(ctx *zero.Ctx, ms []messager) {
	var text string
	for i, m := range ms { // 每4个一组发送
		if i%int(proxy.GetConfigInt64("group")) == 0 && len(text) > 0 {
			ctx.Send(text)
			time.Sleep(100 * time.Millisecond)
			text = ""
		}
		if len(text) > 0 {
			text += "\n\n"
		}
		text += m.GenText(i + 1)
	}
	if len(text) > 0 {
		ctx.Send(text)
	}
}

func (u UserInfo) GenText(index int) string {
	str := u.Name + "\n"
	if index > 0 {
		str = "[" + fmt.Sprintf("%d", index) + "] " + str
	}
	str += "粉丝数：" + fmt.Sprintf("%d", u.Fans) + "\n"
	str += "等级：lv" + fmt.Sprintf("%d", u.Level)
	return str
}

func (u UserInfo) GenMessage(index int) message.Message {
	return message.Message{message.Text(u.GenText(index))}
}

func (b BangumiInfo) GenText(index int) string {
	str := b.Title + "\n"
	if index > 0 {
		str = "[" + fmt.Sprintf("%d", index) + "] " + str
	}
	str += "番剧ID：" + fmt.Sprintf("%d", b.MediaID) + "\n"
	str += "地区：" + b.Areas + "\n"
	if utils.StringRealLength(b.Description) > 50 {
		str += "简介：" + string([]rune(b.Description)[:50]) + "..."
	} else {
		str += "简介：" + b.Description
	}
	return str
}

func (b BangumiInfo) GenMessage(index int) message.Message {
	return message.Message{message.Text(b.GenText(index))}
}
