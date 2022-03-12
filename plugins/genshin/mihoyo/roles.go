package mihoyo

import (
	"encoding/json"
	"errors"
)

type GameRoleList struct {
	List []GameRole `json:"list"`
}
type GameRole struct {
	Uid        string `json:"game_uid"`
	NickName   string `json:"nickname"`
	Region     string `json:"region"`
	RegionName string `json:"region_name"`
}

func GetUserGameRoles(cookie string) ([]GameRole, error) {
	// 请求
	mr := NewMiyoRequest("https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn")
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/app/community-game-records/index.html?v=6")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("X-Requested-With", "com.mihoyo.hyperion")
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	// 解析
	roleList := GameRoleList{}
	err = json.Unmarshal(data, &roleList)
	if err != nil {
		return nil, err
	}
	if len(roleList.List) == 0 {
		return nil, errors.New("没有找到绑定的角色")
	}
	return roleList.List, nil
}

func GetUserGameRoleByUid(cookie string, uid string) (*GameRole, error) {
	roles, err := GetUserGameRoles(cookie)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].Uid == uid {
			return &roles[i], nil
		}
	}
	return nil, errors.New("没有找到UID为" + uid + "的角色")
}
