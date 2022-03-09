package nickname

import (
	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils"
)

// GetNickname 获取用户昵称
func GetNickname(userID int64, defaultName string) string {
	var user dao.UserSetting
	res := proxy.GetDB().Select("nickname").Take(&user, userID)
	if res.RowsAffected > 0 {
		// 检查昵称长度
		max := int(proxy.GetConfigInt64("max"))
		if utils.StringRealLength(user.Nickname) > max {
			if max > 0 {
				user.Nickname = string([]rune(user.Nickname)[:max])
			} else {
				return defaultName
			}
		}
		// 长度为0，返回默认
		if len(user.Nickname) == 0 {
			return defaultName
		}
		return user.Nickname
	}
	return defaultName
}
