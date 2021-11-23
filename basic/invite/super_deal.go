package invite

import (
	"fmt"
	"image"
	"io"

	"github.com/RicheyJang/PaimengBot/utils"

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
	data := make(map[int64]string)
	info := "所有好友：\n"
	for _, friend := range friends {
		id := friend.Get("user_id").Int()
		if id == ctx.Event.SelfID { // 跳过自己
			continue
		}
		name := friend.Get("nickname").String()
		str := fmt.Sprintf("ID:%v 用户名：%v", id, name)
		nick := nickname.GetNickname(id, "")
		if len(nick) > 0 {
			str += fmt.Sprintf(" 昵称：%v", nick)
		}
		data[id] = str
		info += str + "\n"
	}
	// 生成图片
	w, _ := images.MeasureStringDefault(info, 24, 1.3)
	msg, err := formQQImgResponse(data, w, true)
	if err == nil {
		ctx.SendChain(msg)
		return
	}
	// 形成兜底回包消息
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
	data := make(map[int64]string)
	info := "所有群组：\n"
	for _, group := range groups {
		id := group.Get("group_id").Int()
		name := group.Get("group_name").String()
		num := group.Get("member_count").Int()
		str := fmt.Sprintf("ID:%v 群名：%v (%v)", id, name, num)
		data[id] = str
		info += str + "\n"
	}
	// 生成图片
	w, _ := images.MeasureStringDefault(info, 24, 1.3)
	msg, err := formQQImgResponse(data, w, false)
	if err == nil {
		ctx.SendChain(msg)
		return
	}
	// 形成兜底回包消息
	ctx.SendChain(formResponse(info))
}

func formQQImgResponse(data map[int64]string, w float64, isFriend bool) (msg message.MessageSegment, err error) {
	var avaReader io.Reader
	avaSize, fontSize, height := 100, 24.0, 10
	img := images.NewImageCtxWithBGRGBA255(int(w)+avaSize+30, len(data)*(avaSize+20)+30, 255, 255, 255, 255)
	for id, str := range data {
		if isFriend {
			avaReader, err = utils.GetQQAvatar(id, avaSize)
		} else {
			avaReader, err = utils.GetQQGroupAvatar(id, avaSize)
		}
		if err != nil {
			return msg, err
		}
		ava, _, err := image.Decode(avaReader)
		ava = images.ClipImgToCircle(ava)
		if err != nil {
			log.Errorf("Decode avatar err: %v", err)
			return msg, err
		}
		img.DrawImage(ava, 10, height)
		err = img.PasteStringDefault(str, fontSize, 1.3, float64(10+avaSize+10), float64(height+25), w)
		if err != nil {
			return msg, err
		}
		height += avaSize + 20
	}
	file, err := img.SaveTempDefault()
	if err != nil {
		return msg, err
	}
	return message.Image("file:///" + file), nil
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
