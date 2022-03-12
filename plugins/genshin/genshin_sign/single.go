package genshin_sign

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils/images"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func singleSignHandler(ctx *zero.Ctx) {
	user_uid, user_cookie, cookie_msg, err := mihoyo.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookie_msg))
		return
	}
	msg, err := Sign(user_uid, user_cookie)
	if err != nil {
		ctx.Send(images.GenStringMsg(msg))
	}
	ctx.Send(message.Text(fmt.Sprintf("签到:%s", msg)))
}
