package invite

import (
	"fmt"
	"image"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/basic/nickname"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "好友群组管理",
	Usage: `
用法：
	查看所有好友[请求]?
	查看所有群组[请求]?
	同意/拒绝好友请求 [XXX]+
	同意/拒绝群组请求 [XXX]+
	退群 [XXX]+
`,
	IsSuperOnly: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"查看所有好友"}).SetBlock(true).FirstPriority().Handle(handleAllFriends)
	proxy.OnCommands([]string{"查看所有群组", "查看所有群"}).SetBlock(true).FirstPriority().Handle(handleAllGroups)
	proxy.OnRegex("(同意|拒绝)好友(请求)?(.+)").SetBlock(true).FirstPriority().Handle(setFriendRequest)
	proxy.OnRegex("(同意|拒绝)(群|群组)(请求|邀请)?(.+)").SetBlock(true).FirstPriority().Handle(setGroupRequest)
	proxy.OnCommands([]string{"退群"}, zero.OnlyToMe).SetBlock(true).FirstPriority().Handle(quitGroup)
}

func setFriendRequest(ctx *zero.Ctx) {
	// 初始化
	reg := regexp.MustCompile("(同意|拒绝)好友(请求)?(.+)")
	sub := reg.FindStringSubmatch(ctx.MessageString())
	if len(sub) <= 3 {
		ctx.Send("谁嘛？")
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(sub[3]), 10, 64)
	if err != nil {
		ctx.Send("格式错误了哦")
		return
	}
	var userS dao.UserSetting
	if res := proxy.GetDB().Take(&userS, id); res.RowsAffected == 0 || len(userS.Flag) == 0 {
		ctx.Send("脑袋里没有这个人的好友请求哦，或者你可以试试手动添加") // 数据库里没有（没捕获到事件）
		return
	}
	approve := false
	if sub[1] == "同意" {
		approve = true
	}
	flag := userS.Flag
	if approve { // 同意 -> 将flag清空
		if err = proxy.GetDB().Model(&userS).Update("flag", "").Error; err != nil {
			log.Errorf("更新数据库表项(UserSetting)失败，err: %v", err)
		}
	} else { // 拒绝 -> 删除请求
		if err = proxy.GetDB().Delete(&userS, id).Error; err != nil {
			log.Errorf("删除数据库表项(UserSetting)失败，err: %v", err)
		}
	}
	// 设置请求
	ctx.SetFriendAddRequest(flag, approve, "")
	ctx.Send("好哒")
}

func setGroupRequest(ctx *zero.Ctx) {
	// 初始化
	reg := regexp.MustCompile("(同意|拒绝)(群|群组)(请求|邀请)?(.+)")
	sub := reg.FindStringSubmatch(ctx.MessageString())
	if len(sub) <= 4 {
		ctx.Send("谁嘛？")
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(sub[4]), 10, 64)
	if err != nil {
		ctx.Send("格式错误了哦")
		return
	}
	var groupS dao.GroupSetting
	if res := proxy.GetDB().Take(&groupS, id); res.RowsAffected == 0 || len(groupS.Flag) == 0 {
		ctx.Send("脑袋里没有这个群的邀请哦，或者你可以试试手动加入") // 数据库里没有（没捕获到事件）
		return
	}
	approve := false
	if sub[1] == "同意" {
		approve = true
	}
	flag := groupS.Flag
	// 更新数据库
	if approve { // 同意 -> 将flag清空，could_add = true
		if err = proxy.GetDB().Model(&groupS).Updates(map[string]interface{}{"flag": "", "could_add": true}).Error; err != nil {
			log.Errorf("更新数据库表项(GroupSetting)失败，err: %v", err)
		}
	} else { // 拒绝 -> 删除请求
		if err = proxy.GetDB().Delete(&groupS, id).Error; err != nil {
			log.Errorf("删除数据库表项(GroupSetting)失败，err: %v", err)
		}
	}
	// 设置请求
	ctx.SetGroupAddRequest(flag, "invite", approve, "")
	ctx.Send("好哒")
}

func quitGroup(ctx *zero.Ctx) {
	arg := utils.GetArgs(ctx)
	id, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
	if err != nil || id == 0 {
		ctx.Send("格式错误了哦")
		return
	}
	if err = proxy.GetDB().Delete(&dao.GroupSetting{}, id).Error; err != nil {
		log.Errorf("删除数据库表项(GroupSetting)失败，err: %v", err)
	}
	ctx.SetGroupLeave(id, false)
	ctx.Send(fmt.Sprintf("已退出群聊%v", id))
}

func handleAllFriends(ctx *zero.Ctx) {
	// 生成所有好友信息
	data, least := formAllFriends(ctx)
	if arg := utils.GetArgs(ctx); strings.Contains(arg, "请求") {
		data, least = formAllFriendRequest(data, ctx)
		if len(data) == 0 {
			ctx.Send("暂时没有好友请求哦")
			return
		}
	}
	// 生成图片
	w, _ := images.MeasureStringDefault(least, 24, 1.3)
	msg, err := formQQImgResponse(data, w, true)
	if err == nil {
		ctx.SendChain(msg)
		return
	}
	// 形成兜底回包消息
	ctx.SendChain(formResponse(least))
}

func handleAllGroups(ctx *zero.Ctx) {
	// 生成所有群组信息
	data, least := formAllGroups(ctx)
	if arg := utils.GetArgs(ctx); strings.Contains(arg, "请求") || strings.Contains(arg, "邀请") {
		data, least = formAllGroupRequest(data, ctx)
		if len(data) == 0 {
			ctx.Send("暂时没有群邀请哦")
			return
		}
	}
	if len(data) == 0 {
		ctx.Send("暂时没有加入群组哦")
		return
	}
	// 生成图片
	w, _ := images.MeasureStringDefault(least, 24, 1.3)
	msg, err := formQQImgResponse(data, w, false)
	if err == nil {
		ctx.SendChain(msg)
		return
	}
	// 形成兜底回包消息
	ctx.SendChain(formResponse(least))
}

func formAllFriends(ctx *zero.Ctx) (map[int64]string, string) {
	res := ctx.GetFriendList()
	friends := res.Array()
	data := make(map[int64]string)
	least := "所有好友：\n"
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
		least += str + "\n"
	}
	return data, least
}

func formAllFriendRequest(has map[int64]string, ctx *zero.Ctx) (map[int64]string, string) {
	var users []dao.UserSetting
	proxy.GetDB().Where("flag <> ?", "").Find(&users)
	log.Infof("数据库中共%v条好友请求", len(users))
	data := make(map[int64]string)
	least := "所有好友请求：\n"
	for _, user := range users {
		if _, ok := has[user.ID]; ok || len(user.Flag) == 0 { // 好友已添加
			continue
		}
		str := fmt.Sprintf("ID:%v", user.ID)
		data[user.ID] = str
		least += str + "\n"
	}
	return data, least
}

func formAllGroups(ctx *zero.Ctx) (map[int64]string, string) {
	res := ctx.GetGroupList()
	groups := res.Array()
	data := make(map[int64]string)
	least := "所有群组：\n"
	for _, group := range groups {
		id := group.Get("group_id").Int()
		name := group.Get("group_name").String()
		num := group.Get("member_count").Int()
		str := fmt.Sprintf("ID:%v 群名：%v (%v)", id, name, num)
		data[id] = str
		least += str + "\n"
	}
	return data, least
}

func formAllGroupRequest(has map[int64]string, ctx *zero.Ctx) (map[int64]string, string) {
	var groups []dao.GroupSetting
	proxy.GetDB().Where("flag <> ?", "").Find(&groups)
	log.Infof("数据库中共%v条群邀请", len(groups))
	data := make(map[int64]string)
	least := "所有群邀请：\n"
	for _, group := range groups {
		if _, ok := has[group.ID]; ok || len(group.Flag) == 0 { // 好友已添加
			continue
		}
		str := fmt.Sprintf("ID:%v", group.ID)
		data[group.ID] = str
		least += str + "\n"
	}
	return data, least
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
			log.Warnf("Decode avatar err: %v", err)
			return msg, err
		}
		img.DrawImage(ava, 10, height)
		err = img.PasteStringDefault(str, fontSize, 1.3, float64(10+avaSize+10), float64(height+25), w)
		if err != nil {
			return msg, err
		}
		height += avaSize + 20
	}
	imgMsg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("生成图片失败, err: %v", err)
		return msg, err
	}
	return imgMsg, nil
}

func formResponse(info string) message.MessageSegment {
	w, h := images.MeasureStringDefault(info, 16, 1.3)
	img := images.NewImageCtxWithBGRGBA255(int(w)+20, int(h), 255, 255, 255, 255)
	err := img.PasteStringDefault(info, 16, 1.3, 10, 0, w)
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Text(info)
	}
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("生成图片失败, err: %v", err)
		return message.Text(info)
	}
	return msg
}
