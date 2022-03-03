package genshin

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
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
	mr := NewMiyoRequest("https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn")
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/app/community-game-records/index.html?v=6")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("X-Requested-With", "com.mihoyo.hyperion")
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	roleList := GameRoleList{}
	json.Unmarshal(data, &roleList)
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

type GenshinDailyNote struct {
	CurrentResin         int                  `json:"current_resin"`
	MaxResin             int                  `json:"max_resin"`
	ResinRecoveryTime    string               `json:"resin_recovery_time"`
	CurrentExpeditionNum int                  `json:"current_expedition_num"`
	MaxExpeditionNum     int                  `json:"max_expedition_num"`
	Expeditions          []GameRoleExpedition `json:"expeditions"`
	CurrentHomeCoin      int                  `json:"current_home_coin"`
	MaxHomeCoin          int                  `json:"max_home_coin"`
	HomeCoinRecoveryTime string               `json:"home_coin_recovery_time"`
}
type GameRoleExpedition struct {
	AvatarSideIconLink string `json:"avatar_side_icon"`
	Status             string `json:"status"`
	RemainedTime       string `json:"remained_time"`
}

func GetGenshinDailyNote(cookie, uid, server string) (*GenshinDailyNote, error) {
	url := fmt.Sprintf("https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/dailyNote?server=%s&role_id=%s", server, uid)
	mr := NewMiyoRequest(url)
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/app/community-game-records/index.html?v=6")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("X-Requested-With", "com.mihoyo.hyperion")
	appType, appVersion, DS := GetDS(mr.url, "")
	mr.SetHeader("x-rpc-client_type", appType)
	mr.SetHeader("x-rpc-app_version", appVersion)
	mr.SetHeader("DS", DS)
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	dailyNote := GenshinDailyNote{}
	json.Unmarshal(data, &dailyNote)
	return &dailyNote, nil
}

/* contributor: https://github.com/Azure99/GenshinPlayerQuery/issues/20 @lulu666lulu */
func GetDS(url, data string) (t, v, ds string) {
	// 5: mobile web
	const appType = "5"
	const apiSalt = "xV8v4Qu54lUKrEYFZkJhB8cuOh9Asafs"
	const appVersion = "2.11.1"
	rand.Seed(time.Now().UnixNano())

	// unix timestamp
	i := time.Now().Unix()
	// random number 100k - 200k
	r := rand.Intn(100000) + 100000
	// body
	b := data
	// query
	q := func() string {
		urlParts := strings.Split(url, "?")
		if len(urlParts) == 2 {
			querys := strings.Split(urlParts[1], "&")
			sort.Strings(querys)
			return strings.Join(querys, "&")
		}
		return ""
	}()
	c := toHexDigest(fmt.Sprintf("salt=%s&t=%d&r=%d&b=%s&q=%s", apiSalt, i, r, b, q))
	ds = fmt.Sprintf("%d,%d,%s", i, r, c)

	return appType, appVersion, ds
}
func toHexDigest(str string) string {
	h := md5.New()
	h.Write([]byte(str))

	return hex.EncodeToString(h.Sum(nil))
}
