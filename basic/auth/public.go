package auth

import (
	"fmt"
	"math"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm/clause"
)

// CheckPriority 检查权限
func CheckPriority(ctx *zero.Ctx, priority int, tip bool) bool {
	if ctx.Event == nil || ctx.Event.UserID == 0 {
		return false
	}
	level := GetGroupUserPriority(ctx.Event.GroupID, ctx.Event.UserID)
	if utils.IsGroupAnonymous(ctx) { // 匿名消息，权限设为最低
		level = math.MaxInt
	}
	if level > priority {
		if !tip { // 不需要提示
			return false
		}
		if level == math.MaxInt {
			ctx.Send(fmt.Sprintf("你的权限不足喔，需要权限%v", priority))
		} else {
			ctx.Send(fmt.Sprintf("你的权限(%v)不足喔，需要权限%v", level, priority))
		}
		return false
	}
	return true
}

// InitialGroupPriority 初始化指定群的所有管理员权限等级（若不存在）
func InitialGroupPriority(ctx *zero.Ctx, groupID int64) error {
	if ctx == nil || groupID == 0 {
		return fmt.Errorf("wrong param ctx=%v,groupID=%v", ctx, groupID)
	}
	level := proxy.GetConfigInt64("defaultLevel")
	members := ctx.GetGroupMemberList(groupID).Array()
	for _, member := range members {
		if member.Get("role").String() != "owner" && member.Get("role").String() != "admin" {
			continue
		}
		userP := dao.UserPriority{
			ID:       member.Get("user_id").Int(),
			GroupID:  groupID,
			Priority: int(level),
		}
		proxy.GetDB().Clauses(clause.OnConflict{DoNothing: true}).Create(&userP)
	}
	return nil
}

// SetGroupUserPriority 将指定群指定用户的权限等级设为level
func SetGroupUserPriority(groupID, userID int64, level int) error {
	userP := dao.UserPriority{
		ID:       userID,
		GroupID:  groupID,
		Priority: level,
	}
	return proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}, {Name: "group_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"priority"}),
	}).Create(&userP).Error
}

// GetGroupUserPriority 获取指定群指定用户的权限等级，数字越小，代表权限越高
func GetGroupUserPriority(groupID, userID int64) (level int) {
	defer func() {
		if utils.IsSuperUser(userID) { // 超级用户单独处理
			sl := int(proxy.GetConfigInt64("superLevel"))
			if sl > 0 && sl < level { // 若默认超级用户拥有更高权限
				level = sl
			}
		}
	}()
	var userP dao.UserPriority
	res := proxy.GetDB().Where(&dao.UserPriority{ID: userID, GroupID: groupID}).Order("priority desc").Limit(1).Find(&userP)
	if res.RowsAffected == 0 || res.Error != nil {
		return math.MaxInt
	}
	return userP.Priority
}
