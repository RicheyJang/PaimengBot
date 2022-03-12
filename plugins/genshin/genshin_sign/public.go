package genshin_sign

import (
	"encoding/json"
	"fmt"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"

	log "github.com/sirupsen/logrus"
)

func Sign(uid string, cookie string) (string, error) {
	// 获取角色信息
	gameRole, _ := mihoyo.GetUserGameRoleByUid(cookie, uid)
	// 米游社签到
	if err := mihoyo.Sign(cookie, *gameRole); err != nil { // 签到失败
		msg := fmt.Sprintf("UID:%v, 昵称:%v: 米游社签到失败",
			gameRole.Uid, gameRole.NickName)
		return msg, err
	}
	// 签到成功
	msg := fmt.Sprintf("UID:%v, 昵称:%v: 米游社签到成功", gameRole.Uid, gameRole.NickName)
	// 查询签到信息
	data, err := mihoyo.GetSignStateInfo(cookie, *gameRole)
	if err != nil {
		log.Errorf("GetSignStateInfo err: %v", err)
		return msg, nil
	}
	msg += fmt.Sprintf(", 已连续签到%v天", data.TotalSignDay)
	// 查询奖励信息
	awards, err := mihoyo.GetSignAwardsList()
	if err != nil {
		log.Errorf("GetSignStateInfo err: %v", err)
		return msg, nil
	}
	if len(awards.Awards) >= data.TotalSignDay {
		item := awards.Awards[data.TotalSignDay-1]
		msg += fmt.Sprintf(", 今天获得%d个%v", item.Count, item.Name)
	}
	return msg, nil
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
