package bilibili

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/wdvxdr1123/ZeroBot/message"
)

func (u UserInfo) GenMessage(index int) message.Message {
	str := u.Name + "\n"
	if index > 0 {
		str = "[" + fmt.Sprintf("%d", index) + "] " + str
	}
	str += "粉丝数：" + fmt.Sprintf("%d", u.Fans) + "\n"
	str += "等级：lv" + fmt.Sprintf("%d", u.Level)
	return message.Message{message.Text(str)}
}

func (b BangumiInfo) GenMessage(index int) message.Message {
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
	return message.Message{message.Text(str)}
}
