package genshin_sign

import (
	"fmt"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_client"

	"time"
)

func Sign(uid string, cookie string) (string, error) {

	g := sign_client.NewGenshinClient()

	gameRolesList := g.GetUserGameRoles(cookie)

	for j := 0; j < len(gameRolesList); j++ {
		time.Sleep(10 * time.Second)
		msg := ""
		if g.Sign(cookie, gameRolesList[j]) {
			time.Sleep(10 * time.Second)
			data := g.GetSignStateInfo(cookie, gameRolesList[j])
			msg = fmt.Sprintf("UID:%v, 昵称:%v, 连续签到天数:%v. 签到成功.",
				gameRolesList[j].UID, gameRolesList[j].Name, data.TotalSignDay)
		} else {
			msg = fmt.Sprintf("UID:%v, 昵称:%v. 签到失败.",
				gameRolesList[j].UID, gameRolesList[j].Name)
		}
		return msg, nil
	}
	return "未知错误", nil

}
