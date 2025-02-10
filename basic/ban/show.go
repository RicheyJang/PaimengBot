package ban

import (
	"fmt"
	"strings"

	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/RicheyJang/PaimengBot/manager"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// userID -> 被封插件列表(白名单为启用的列表)
func getUsersBlackOrWhite(filterWhite bool) map[int64][]string {
	var users []dao.UserSetting

	proxy.GetDB().Select("id", "black_plugins", "white_plugins", "white_mode").Where("LENGTH(black_plugins) > 1 or LENGTH(white_plugins) > 1").Find(&users)
	res := make(map[int64][]string)
	for _, user := range users {
		var plugins []string
		if user.WhiteMode != filterWhite {
			continue
		}
		if filterWhite {
			plugins = utils.MergeStringSlices(strings.Split(user.WhitePlugins, "|")) // 去重去空
		} else {
			plugins = utils.MergeStringSlices(strings.Split(user.BlackPlugins, "|")) // 去重去空
		}
		if len(plugins) == 0 {
			continue
		}
		res[user.ID] = plugins
	}
	return res
}

// groupID -> 被封插件列表(白名单为启用的列表)
func getGroupsBlackOrWhite(filterWhite bool) map[int64][]string {
	var groups []dao.GroupSetting
	proxy.GetDB().Select("id", "black_plugins", "white_plugins", "white_mode").Where("LENGTH(black_plugins) > 1").Find(&groups)
	res := make(map[int64][]string)
	for _, group := range groups {
		var plugins []string
		if group.WhiteMode != filterWhite {
			continue
		}
		if filterWhite {
			plugins = utils.MergeStringSlices(strings.Split(group.WhitePlugins, "|")) // 去重去空
		} else {
			plugins = utils.MergeStringSlices(strings.Split(group.BlackPlugins, "|")) // 去重去空
		}
		if len(plugins) == 0 {
			continue
		}
		res[group.ID] = plugins
	}
	return res
}

func showBlack(ctx *zero.Ctx) {
	if !utils.IsMessageGroup(ctx) {
		if !utils.IsSuperUser(ctx.Event.UserID) || !utils.IsMessagePrimary(ctx) {
			ctx.Send("请在群聊中问我哦")
			return
		}
		// 超级用户 in 私聊 单独处理
		showBlackInPrimarySuper(ctx)
		return
	}
	// 群管理员 in 群聊
	var str string
	userM := getUsersBlackOrWhite(false)
	rsp := ctx.GetGroupMemberList(ctx.Event.GroupID) // 过滤掉非本群成员
	users := rsp.Array()
	for _, user := range users {
		id := user.Get("user_id").Int()
		if userBlack, ok := userM[id]; ok && id != 0 {
			str += fmt.Sprintf("%v(%v %v): %v\n",
				user.Get("nickname"), user.Get("card"), id, formBlackDescription(userBlack))
		}
	}
	if len(str) == 0 {
		ctx.Send("大家都是好人，黑名单暂时是空哒")
		return
	}
	// 生成图片
	w, h := images.MeasureStringDefault(str, 24, 1.3)
	w, h = w+20, h+20
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	err := img.PasteStringDefault(str, 24, 1.3, 10, 10, w)
	if err != nil {
		ctx.Send(str)
		return
	}
	msg, err := img.GenMessageAuto()
	if err != nil {
		ctx.Send(str)
		return
	}
	ctx.SendChain(msg)
}

func showWhite(ctx *zero.Ctx) {
	userM := getUsersBlackOrWhite(true)
	groupM := getGroupsBlackOrWhite(true)
	if len(userM) == 0 && len(groupM) == 0 {
		ctx.Send("没有群/用户运行在白名单模式")
		return
	}
	// 用户
	if len(userM) > 0 {
		str := "用户："
		userDesM := make(map[int64]string)
		for id, blacks := range userM {
			var des string
			if id == 0 { // 全体用户（全局）
				des = fmt.Sprintf("全体: %v\n", formBlackDescription(blacks))
			} else { // 正常用户
				user := ctx.GetStrangerInfo(id, false)
				des = fmt.Sprintf("%v(%v): %v\n", user.Get("nickname"), id, formBlackDescription(blacks))
			}
			userDesM[id] = des
			str += des + "\n"
		}
		w, _ := images.MeasureStringDefault(str, 24, 1.3)
		msg, err := images.GenQQListMsgWithAva(userDesM, w, true)
		if err != nil {
			ctx.Send(str)
		} else {
			ctx.SendChain(message.Text("用户：\n"), msg)
		}
	}
	// 群聊
	if len(groupM) > 0 {
		str := "群："
		groupDesM := make(map[int64]string)
		for id, blacks := range groupM {
			group := ctx.GetGroupInfo(id, false)
			des := fmt.Sprintf("%v(%v): %v\n",
				group.Name, id, formBlackDescription(blacks))
			groupDesM[id] = des
			str += des + "\n"
		}
		w, _ := images.MeasureStringDefault(str, 24, 1.3)
		msg, err := images.GenQQListMsgWithAva(groupDesM, w, false)
		if err != nil {
			ctx.Send(str)
		} else {
			ctx.SendChain(message.Text("群：\n"), msg)
		}
	}
}

func showBlackInPrimarySuper(ctx *zero.Ctx) {
	userM := getUsersBlackOrWhite(false)
	groupM := getGroupsBlackOrWhite(false)
	if len(userM) == 0 && len(groupM) == 0 {
		ctx.Send("大家都是好人，黑名单暂时是空哒")
		return
	}
	// 用户
	if len(userM) > 0 {
		str := "用户："
		userDesM := make(map[int64]string)
		for id, blacks := range userM {
			var des string
			if id == 0 { // 全体用户（全局）
				des = fmt.Sprintf("全体: %v\n", formBlackDescription(blacks))
			} else { // 正常用户
				user := ctx.GetStrangerInfo(id, false)
				des = fmt.Sprintf("%v(%v): %v\n", user.Get("nickname"), id, formBlackDescription(blacks))
			}
			userDesM[id] = des
			str += des + "\n"
		}
		w, _ := images.MeasureStringDefault(str, 24, 1.3)
		msg, err := images.GenQQListMsgWithAva(userDesM, w, true)
		if err != nil {
			ctx.Send(str)
		} else {
			ctx.SendChain(message.Text("用户：\n"), msg)
		}
	}
	// 群聊
	if len(groupM) > 0 {
		str := "群："
		groupDesM := make(map[int64]string)
		for id, blacks := range groupM {
			group := ctx.GetGroupInfo(id, false)
			des := fmt.Sprintf("%v(%v): %v\n",
				group.Name, id, formBlackDescription(blacks))
			groupDesM[id] = des
			str += des + "\n"
		}
		w, _ := images.MeasureStringDefault(str, 24, 1.3)
		msg, err := images.GenQQListMsgWithAva(groupDesM, w, false)
		if err != nil {
			ctx.Send(str)
		} else {
			ctx.SendChain(message.Text("群：\n"), msg)
		}
	}
}

func formBlackDescription(blacks []string) string {
	var des string
	for i, black := range blacks {
		if black == AllPluginKey {
			des = "全部功能"
			blacks = append(blacks[:i], blacks[i+1:]...)
			break
		}
	}
	if len(blacks) > 0 && len(des) > 0 {
		des += "+"
	}
	for _, black := range blacks {
		plugin := manager.GetPluginConditionByKey(black)
		if plugin == nil {
			continue
		}
		if len(des) > 0 && des[0] != '+' {
			des += "、"
		}
		des += plugin.Name
	}
	return des
}
