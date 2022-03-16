package admin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/ban"
	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm/clause"
)

func blackSomeone(ctx *zero.Ctx) {
	if !utils.IsMessageGroup(ctx) && !utils.IsSuperUser(ctx.Event.UserID) {
		ctx.Send("请在群聊中调用此功能")
		return
	}
	id, _, err := analysisArgs(ctx, false)
	if err != nil {
		log.Errorf("解析参数错误：%v", err)
		return
	}
	if id <= 0 {
		ctx.Send("请指定拉黑的QQ号或@")
		return
	}
	// 禁止拉黑机器人自身和超级用户
	if id == ctx.Event.SelfID || utils.IsSuperUser(id) {
		ctx.Send("？")
		return
	}
	// 获取确认
	ctx.Send(fmt.Sprintf("确认拉黑用户%v？这将踢出并自动拒绝该用户加入任何%[2]v作为管理员的群聊，删除并禁止其加%[2]v为好友，封禁该用户的所有功能使用权", id, utils.GetBotNickname()))
	event := utils.WaitNextMessage(ctx)
	if event == nil { // 无回应
		return
	}
	confirm := strings.TrimSpace(event.Message.ExtractPlainText())
	if !(confirm == "是" || confirm == "确定" || confirm == "确认") {
		ctx.Send("已取消")
		return
	}
	// 修改IsPullBlack
	preUser := dao.UserSetting{
		ID:          id,
		IsPullBlack: true,
	}
	if err = proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_pull_black"}), // Upsert
	}).Create(&preUser).Error; err != nil {
		log.Errorf("set user(%v) is_pull_black error(sql): %v", id, err)
		ctx.Send("失败了...")
		return
	}
	// 踢出所有群
	groups := ctx.GetGroupList().Array()
	for _, group := range groups {
		groupID := group.Get("group_id").Int()
		if groupID == 0 {
			continue
		}
		black := ctx.GetGroupMemberInfo(groupID, id, false)
		if black.Get("role").String() != "member" { // 拉黑成员不在该群或无法踢出
			continue
		}
		self := ctx.GetGroupMemberInfo(groupID, ctx.Event.SelfID, false)
		if self.Get("role").String() == "member" { // 机器人并非管理员
			log.Warnf("该用户在群%d中，但%v没有管理员权限，在群%[1]d中无法踢出该用户", groupID, utils.GetBotNickname())
			continue
		}
		ctx.SetGroupKick(groupID, id, false)
	}
	// 删好友
	ctx.CallAction("delete_friend", zero.Params{
		"friend_id": id,
	})
	// 封禁
	_ = ban.SetUserPluginStatus(false, id, nil, 0)
	ctx.Send("走你")
}

func unBlackSomeone(ctx *zero.Ctx) {
	if !utils.IsMessageGroup(ctx) && !utils.IsSuperUser(ctx.Event.UserID) {
		ctx.Send("请在群聊中调用此功能")
		return
	}
	id, _, err := analysisArgs(ctx, false)
	if err != nil {
		log.Errorf("解析参数错误：%v", err)
		return
	}
	if id <= 0 {
		ctx.Send("请指定取消拉黑的QQ号或@")
		return
	}
	// 修改IsPullBlack
	if err = proxy.GetDB().Model(&dao.UserSetting{ID: id}).Update("is_pull_black", false).Error; err != nil {
		log.Errorf("set user(%v) is_pull_black error(sql): %v", id, err)
		ctx.Send("失败了...")
		return
	}
	// 解封
	_ = ban.SetUserPluginStatus(true, id, nil, 0)
	ctx.Send("好哒")
}

func blackList(ctx *zero.Ctx) {
	var users []dao.UserSetting
	proxy.GetDB().Where(&dao.UserSetting{IsPullBlack: true}).Find(&users)
	if len(users) == 0 {
		ctx.Send("暂时没有用户被拉黑")
		return
	}
	// 生成消息
	userMap := make(map[int64]string)
	least := "被拉黑用户清单："
	for _, user := range users {
		id := "QQ: " + strconv.FormatInt(user.ID, 10)
		userMap[user.ID] = id
		least += "\n" + id
	}
	w, _ := images.MeasureStringDefault(least, 24, 1.3)
	msg, err := images.GenQQListMsgWithAva(userMap, w, true)
	if err != nil {
		log.Warnf("GenQQListMsgWithAva err: %v", err)
		ctx.Send(least)
	} else {
		ctx.Send(msg)
	}
}
