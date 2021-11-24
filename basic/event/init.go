package event

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name:     "处理除消息外其它基本事件",
	Usage:    "防止被动拉入群聊；捕获好友、群邀请发送给超级用户",
	IsHidden: true,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRequest().FirstPriority().Handle(handleInvite) // 捕获好友、群邀请发送给超级用户
	proxy.OnNotice(utils.CheckDetailType("group_increase"), func(ctx *zero.Ctx) bool {
		return ctx.Event.SelfID == ctx.Event.UserID
	}).SetBlock(true).FirstPriority().Handle(preventForcedInviteGroup) // 防止被动拉入群聊
	proxy.AddConfig("notAutoLeave", false)
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
	utils.SendToSuper(message.Text(fmt.Sprintf("%v被%v拉入了群%v",
		utils.GetBotNickname(), ctx.Event.OperatorID, ctx.Event.GroupID)))
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
	tmpUser := &dao.UserSetting{}
	if res := proxy.GetDB().Where(&userS, "id", "flag").Find(tmpUser); res.RowsAffected > 0 {
		return
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"flag"}), // Upsert
	}).Create(&userS).Error; err != nil {
		log.Errorf("set user(id=%v) flag error(sql): %v", ctx.Event.UserID, err)
		utils.SendToSuper(message.Text("处理好友请求时SQL出错，请尽快处理"))
	} else {
		str := fmt.Sprintf("收到一条好友请求：\nID: %v\n验证消息：%v", ctx.Event.UserID, ctx.Event.Comment)
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
		utils.SendToSuper(message.Text("处理群邀请请求时SQL出错，请尽快处理"))
	} else {
		str := fmt.Sprintf("收到一条群邀请：\nID: %v\n邀请者ID：%v", ctx.Event.GroupID, ctx.Event.UserID)
		utils.SendToSuper(message.Text(str))
	}
}
