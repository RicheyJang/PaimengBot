package mihoyo

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

type SignState struct {
	Today        string `json:"today"`
	TotalSignDay int    `json:"total_sign_day"`
	IsSign       bool   `json:"is_sign"`
	IsSub        bool   `json:"is_sub"`
	MonthFirst   bool   `json:"month_first"`
	FirstBind    bool   `json:"first_bind"`
}

type SignAwardsList struct {
	Awards []struct {
		Name  string `json:"name"`
		Count int    `json:"cnt"`
	} `json:"awards"`
}

func Sign(cookie string, user GameRole) error {
	data := map[string]string{
		"act_id": "e202009291139501",
		"region": user.Region,
		"uid":    user.Uid,
	}
	mr := NewMiyoRequest("https://api-takumi.mihoyo.com/event/bbs_sign_reward/sign")
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Linux; Android 5.1.1; f103 Build/LYZ28N; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 miHoYoBBS/2.3.0")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/bbs/event/signin-ys/index.html?bbs_auth_required=true&act_id=e202009291139501&utm_source=bbs&utm_medium=mys&utm_campaign=icon")
	mr.SetHeader("Accept-Encoding", "gzip, deflate")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("x-rpc-device_id", uuid.NewV4().String())
	appType, appVersion, DS := getSignDS()
	mr.SetHeader("x-rpc-client_type", appType)
	mr.SetHeader("x-rpc-app_version", appVersion)
	mr.SetHeader("DS", DS)
	// 解析
	rsp, err := mr.Post(data)
	if err != nil {
		log.Errorf("Miyo Sign Post err: %v", err)
		return err
	}
	if rsp.RetCode == 0 || rsp.RetCode == -5003 {
		return nil
	}
	log.Errorf("Miyo Sign illegal response: %+v", rsp)
	return fmt.Errorf("sign response code=%v, message=%v", rsp.RetCode, rsp.Message)
}

func GetSignStateInfo(cookie string, user GameRole) (*SignState, error) {
	// 请求
	url := fmt.Sprintf("https://api-takumi.mihoyo.com/event/bbs_sign_reward/info?act_id=%s&region=%s&uid=%s", "e202009291139501", user.Region, user.Uid)
	mr := NewMiyoRequest(url)
	mr.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/app/community-game-records/index.html?v=6")
	mr.SetHeader("Cookie", cookie)
	mr.SetHeader("X-Requested-With", "com.mihoyo.hyperion")
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	// 解析
	signState := SignState{}
	err = json.Unmarshal(data, &signState)
	return &signState, err
}

func GetSignAwardsList() (*SignAwardsList, error) {
	// 请求
	mr := NewMiyoRequest("https://api-takumi.mihoyo.com/event/bbs_sign_reward/home?act_id=e202009291139501")
	mr.SetHeader("x-rpc-app_version", "2.11.1")
	mr.SetHeader("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) miHoYoBBS/2.11.1")
	mr.SetHeader("Referer", "https://webstatic.mihoyo.com/")
	mr.SetHeader("x-rpc-client_type", "5")
	data, err := mr.Execute()
	if err != nil {
		return nil, err
	}
	// 解析
	signState := SignAwardsList{}
	err = json.Unmarshal(data, &signState)
	return &signState, err
}

func getSignDS() (t string, v string, ds string) {
	const appType = "5"
	const apiSalt = "h8w582wxwgqvahcdkpvdhbh2w9casgfl"
	const appVersion = "2.3.0"

	currentTime := time.Now().Unix()
	stringRom := getRandString(6, currentTime)
	stringAdd := fmt.Sprintf("salt=%s&t=%d&r=%s", apiSalt, currentTime, stringRom)
	stringMd5 := toHexDigest(stringAdd)
	return appType, appVersion, fmt.Sprintf("%d,%s,%s", currentTime, stringRom, stringMd5)
}

func getRandString(len int, seed int64) string {
	bytes := make([]byte, len)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < len; i++ {
		b := r.Intn(36)
		if b > 9 {
			b += 39
		}
		b += 48
		bytes[i] = byte(b)
	}
	return string(bytes)
}
