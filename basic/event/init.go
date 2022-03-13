package event

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/rules"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "基本事件处理",
	Usage: `防止被动拉入群聊；捕获好友、群邀请发送给超级用户
config-plugin配置项：
	event.notautoleave: 是(true)否(false)关闭被动拉群时自动退群
	event.autoagree: 是(true)否(false)自动同意所有好友请求`,
	IsPassive:   true,
	IsSuperOnly: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRequest().FirstPriority().Handle(handleInvite) // 捕获好友、群邀请发送给超级用户
	proxy.OnNotice(rules.CheckDetailType("group_increase"), func(ctx *zero.Ctx) bool {
		return ctx.Event.SelfID == ctx.Event.UserID
	}).SetBlock(true).FirstPriority().Handle(preventForcedInviteGroup) // 防止被动拉入群聊
	proxy.OnNotice(rules.CheckDetailType("group_admin")).FirstPriority().Handle(handleGroupAdmin)
	proxy.AddConfig("notAutoLeave", false)
	proxy.AddConfig("autoAgree", false)
}

// 机器人初入群聊时
func preventForcedInviteGroup(ctx *zero.Ctx) {
	if utils.IsSuperUser(ctx.Event.OperatorID) { // 超级用户操作
		return
	}
	var groupS dao.GroupSetting
	res := proxy.GetDB().Take(&groupS, ctx.Event.GroupID)
	if !utils.IsSuperUser(ctx.Event.OperatorID) &&
		(res.RowsAffected == 0 || (len(groupS.Flag) != 0 && !groupS.CouldAdd)) &&
		!proxy.GetConfigBool("notAutoLeave") { // 自动退群
		ctx.SendGroupMessage(ctx.Event.GroupID, fmt.Sprintf("请不要随便拉%v入群", utils.GetBotNickname()))
		ctx.SetGroupLeave(ctx.Event.GroupID, false)
		utils.SendToSuper(message.Text(fmt.Sprintf("%v被%v强制拉入了群%v，但是%v及时退群啦",
			utils.GetBotNickname(), ctx.Event.OperatorID, ctx.Event.GroupID, utils.GetBotNickname())))
		return
	}
	// 可以加群 -> 清除flag | 创建记录
	groupS.ID = ctx.Event.GroupID
	groupS.Flag = ""
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"flag"}), // Upsert
	}).Create(&groupS).Error; err != nil {
		log.Errorf("set group(id=%v) flag error(sql): %v", groupS.ID, err)
	}
	utils.SendToSuper(message.Text(fmt.Sprintf("%v成功加入了群%v",
		utils.GetBotNickname(), ctx.Event.GroupID)))
	go auth.InitialGroupPriority(ctx, ctx.Event.GroupID) // 初始化群管理员权限等级
}

// 群管理员变动时
func handleGroupAdmin(ctx *zero.Ctx) {
	if ctx.Event.GroupID == 0 {
		return
	}
	// 重新初始化群权限
	err := auth.InitialGroupPriority(ctx, ctx.Event.GroupID)
	if err != nil {
		log.Warnf("初始化群%v权限失败, err: %v", ctx.Event.GroupID, err)
	} else {
		log.Infof("初始化群%v权限成功", ctx.Event.GroupID)
	}
	// 取消管理员时，收回该管理员权限
	if ctx.Event.SubType == "unset" && ctx.Event.UserID != 0 {
		err = auth.SetGroupUserPriority(ctx.Event.GroupID, ctx.Event.UserID, 0)
		if err != nil {
			log.Warnf("撤除群%v管理员%v权限失败, err: %v", ctx.Event.GroupID, ctx.Event.UserID, err)
		} else {
			log.Warnf("撤除群%v管理员%v权限成功", ctx.Event.GroupID, ctx.Event.UserID)
		}
	}
}

// 收到邀请入群、加好友请求时
func handleInvite(ctx *zero.Ctx) {
	switch ctx.Event.RequestType {
	case "friend":
		handleFriendRequest(ctx)
	case "group":
		if ctx.Event.SubType == "invite" {
			handleGroupInvite(ctx)
		}
	}
}

func handleFriendRequest(ctx *zero.Ctx) {
	userS := dao.UserSetting{
		ID:   ctx.Event.UserID,
		Flag: ctx.Event.Flag,
	}
	str := fmt.Sprintf("收到一条好友请求：\nID: %v\n验证消息：%v", ctx.Event.UserID, ctx.Event.Comment)
	// 自动同意
	if proxy.GetConfigBool("autoAgree") {
		ctx.SetFriendAddRequest(userS.Flag, true, "")
		userS.Flag = ""
		str += "\n根据配置，已自动同意"
	} else {
		str += fmt.Sprintf("\n若同意请说：同意好友请求 %[1]v\n若拒绝请说：拒绝好友请求 %[1]v\n", ctx.Event.UserID)
	}
	// 正常处理
	if res := proxy.GetDB().Where(&userS, "id", "flag").Take(&dao.UserSetting{}); res.RowsAffected > 0 {
		if len(userS.Flag) == 0 { // 自动同意消息
			utils.SendToSuper(message.Text(str))
		}
		return
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"flag"}), // Upsert
	}).Create(&userS).Error; err != nil {
		log.Errorf("set user(id=%v) flag error(sql): %v", ctx.Event.UserID, err)
		utils.SendToSuper(message.Text("处理好友请求时SQL出错，请尽快查看日志处理"))
	} else {
		utils.SendToSuper(message.Text(str))
	}
}

func handleGroupInvite(ctx *zero.Ctx) {
	groupS := dao.GroupSetting{
		ID:   ctx.Event.GroupID,
		Flag: ctx.Event.Flag,
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"flag"}), // Upsert
	}).Create(&groupS).Error; err != nil {
		log.Errorf("set group(id=%v) flag error(sql): %v", ctx.Event.GroupID, err)
		utils.SendToSuper(message.Text("处理群邀请请求时SQL出错，请尽快查看日志处理"))
	} else {
		str := fmt.Sprintf("收到一条群邀请：\n群ID: %v\n邀请者ID：%v", ctx.Event.GroupID, ctx.Event.UserID)
		str += fmt.Sprintf("\n若同意请说：同意群邀请 %[1]v\n若拒绝请说：拒绝群邀请 %[1]v", ctx.Event.GroupID)
		utils.SendToSuper(message.Text(str))
	}
}
