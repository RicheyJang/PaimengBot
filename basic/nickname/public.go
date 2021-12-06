package nickname

import "github.com/RicheyJang/PaimengBot/basic/dao"

// GetNickname 获取用户昵称
func GetNickname(userID int64, defaultName string) string {
	var user dao.UserSetting
	res := proxy.GetDB().Select("nickname").Take(&user, userID)
	if res.RowsAffected > 0 {
		return user.Nickname
	}
	return defaultName
}
