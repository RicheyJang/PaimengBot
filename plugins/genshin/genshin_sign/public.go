package genshin_sign

import (
	"encoding/json"
	"fmt"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign/sign_client"
)

func Sign(uid string, cookie string) (string, error) {
	g := sign_client.NewGenshinClient()
	gameRolesList := g.GetUserGameRoles(cookie)
	for j := 0; j < len(gameRolesList); j++ {
		//time.Sleep(10 * time.Second)
		msg := ""
		if g.Sign(cookie, gameRolesList[j]) {
			//time.Sleep(10 * time.Second)
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

type EventFrom struct {
	IsFromGroup bool
	FromId      string
	QQ          string `json:"qq"`
	Auto        bool
}

type UserInfo struct {
	ID        string
	Uin       string
	cookie    string
	EventFrom EventFrom
}

func GetEventFrom(id int64) (eventFrom EventFrom, e error) {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		e = err
		return
	}
	// 解析
	e = json.Unmarshal(v, &eventFrom)
	return
}

func PutEventFrom(id int64, u EventFrom) error {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}
