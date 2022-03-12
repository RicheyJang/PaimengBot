package genshin_public

import (
	"errors"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_cookie"
)

func GetUidCookieById(id int64) (string, string, string, error) {
	user_cookie := genshin_cookie.GetUserCookie(id)
	user_uid := genshin_cookie.GetUserUid(id)
	cookie_msg := GetInitializaationPrompt()
	if len(user_cookie) <= 10 {
		return "", "", cookie_msg + "\ncookie设置失败", errors.New(cookie_msg + "\ncookie设置失败")
	}
	if len(user_uid) <= 5 {
		return "", "", cookie_msg + "\nuid设置失败", errors.New(cookie_msg + "\nuid设置失败")
	}
	return user_uid, user_cookie, "获取成功", nil
}
