package invite

import (
	"fmt"

	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/RicheyJang/PaimengBot/basic/nickname"
	"github.com/RicheyJang/PaimengBot/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "好友群组管理",
	Usage: `
用法：
	查看所有好友
	查看所有群组
	同意/拒绝好友请求 [XXX]+
	同意/拒绝群组请求 [XXX]+
`,
	IsSuperOnly: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"查看所有好友"}).SetBlock(true).FirstPriority().Handle(handleAllFriends)
	proxy.OnCommands([]string{"查看所有群", "查看所有群组"}).SetBlock(true).FirstPriority().Handle(handleAllGroups)
}

func handleAllFriends(ctx *zero.Ctx) {
	res := ctx.GetFriendList()
	friends := res.Array()
	// 生成所有好友信息
	info := "所有好友：\n"
	for _, friend := range friends {
		id := friend.Get("user_id").Int()
		if id == ctx.Event.SelfID { // 跳过自己
			continue
		}
		name := friend.Get("nickname").String()
		info += fmt.Sprintf("ID:%v 用户名：%v", id, name)
		nick := nickname.GetNickname(id, "")
		if len(nick) > 0 {
			info += fmt.Sprintf(" 昵称：%v\n", nick)
		} else {
			info += "\n"
		}
	}
	// 形成回包消息
	ctx.SendChain(formResponse(info))
}

func handleAllGroups(ctx *zero.Ctx) {
	res := ctx.GetGroupList()
	groups := res.Array()
	if len(groups) == 0 {
		ctx.Send("暂时没有加入群组哦")
		return
	}
	// 生成所有群组信息
	info := "所有群组：\n"
	for _, group := range groups {
		id := group.Get("group_id").Int()
		name := group.Get("group_name").String()
		num := group.Get("member_count").Int()
		info += fmt.Sprintf("ID:%v 群名：%v (%v)\n", id, name, num)
	}
	// 形成回包消息
	ctx.SendChain(formResponse(info))
}

func formResponse(info string) message.MessageSegment {
	w, h := images.MeasureStringDefault(info, 16, 1.3)
	img := images.NewImageCtxWithBGRGBA255(int(w)+20, int(h), 255, 255, 255, 255)
	err := img.PasteStringDefault(info, 16, 1.3, 10, 0, w)
	if err != nil {
		log.Errorf("PasteStringDefault err: %v", err)
		return message.Text(info)
	}
	file, err := img.SaveTempDefault()
	if err != nil {
		return message.Text(info)
	}
	return message.Image("file:///" + file)
}
