package mihoyo

import (
	"errors"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_cookie"
)

func GetUidCookieById(id int64) (string, string, string, error) {
	userCookie := genshin_cookie.GetUserCookie(id)
	userUid := genshin_cookie.GetUserUid(id)
	if len(userCookie) <= 10 {
		return "", "", "cookie设置失败\n" + GetInitializaationPrompt(), errors.New("cookie设置失败")
	}
	if len(userUid) <= 5 {
		return "", "", "uid设置失败\n" + GetInitializaationPrompt(), errors.New("uid设置失败")
	}
	return userUid, userCookie, "获取成功", nil
}
