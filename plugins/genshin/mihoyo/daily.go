package mihoyo

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type GenshinDailyNote struct {
	CurrentResin           int                  `json:"current_resin"`                 // 当前树脂
	MaxResin               int                  `json:"max_resin"`                     // 最大树脂
	ResinRecoveryTime      string               `json:"resin_recovery_time"`           // 树脂恢复剩余时间
	FinishedTaskNum        int                  `json:"finished_task_num"`             // 委托完成数
	TotalTaskNum           int                  `json:"total_task_num"`                // 最大委托数
	GetTaskExReward        bool                 `json:"is_extra_task_reward_received"` // 是否已打扰凯瑟琳
	RemainResinDiscountNum int                  `json:"remain_resin_discount_num"`     // 周本体力减半剩余次数
	ResinDiscountNumLimit  int                  `json:"resin_discount_num_limit"`      // 周本体力减半总次数
	CurrentExpeditionNum   int                  `json:"current_expedition_num"`        // 当前派遣数
	MaxExpeditionNum       int                  `json:"max_expedition_num"`            // 最大派遣数
	Expeditions            []GameRoleExpedition `json:"expeditions"`                   // 派遣角色详情
	CurrentHomeCoin        int                  `json:"current_home_coin"`             // 当前洞天宝钱数
	MaxHomeCoin            int                  `json:"max_home_coin"`                 // 最大洞天宝钱数
	HomeCoinRecoveryTime   string               `json:"home_coin_recovery_time"`       // 洞天宝钱恢复剩余时间
}
type GameRoleExpedition struct {
	AvatarSideIconLink string `json:"avatar_side_icon"` // 派遣中角色侧头像
	Status             string `json:"status"`           // 派遣状态：Finished为完成
	RemainedTime       string `json:"remained_time"`    // 派遣完成剩余时间
}

func GetGenshinDailyNote(cookie, uid, server string) (*GenshinDailyNote, error) {
	// 请求
	url := fmt.Sprintf("https://api-takumi-record.mihoyo.com/game_record/app/genshin/api/dailyNote?server=%s&role_id=%s", server, uid)
	mr := NewMiyoRequest(url)
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/app/community-game-records/index.html?v=6")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("X-Requested-With", "com.mihoyo.hyperion")
	appType, appVersion, DS := getDS(mr.url, "")
	mr.SetHeader("x-rpc-client_type", appType)
	mr.SetHeader("x-rpc-app_version", appVersion)
	mr.SetHeader("DS", DS)
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	// 解析
	dailyNote := GenshinDailyNote{}
	err = json.Unmarshal(data, &dailyNote)
	return &dailyNote, err
}

/* contributor: https://github.com/Azure99/GenshinPlayerQuery/issues/20 @lulu666lulu */
func getDS(url, data string) (t, v, ds string) {
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
