package translate

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/utils/client"

	"github.com/RicheyJang/PaimengBot/utils"
)

// BaiduCheckLangSupport 检查该语种是否支持：若不支持，返回空字符串;若支持，返回纯小写字母语种代号
func BaiduCheckLangSupport(lang string, isFrom bool) (res string) {
	defer func() {
		if res == "auto" && !isFrom { // 在目标语种中使用自动
			res = ""
		}
	}()
	if trans, ok := BaiduTransLangMap[lang]; ok { // key中查找
		return trans
	}
	for _, trans := range BaiduTransLangMap { // val中查找
		if trans == lang {
			return trans
		}
	}
	return ""
}

// BaiduTranslate 使用百度翻译API（每月超过200万字符收费）进行翻译
func BaiduTranslate(str, from, to string) (string, error) {
	// 检查
	if len(proxy.GetConfigString("baidu.appid")) == 0 || len(proxy.GetConfigString("baidu.key")) == 0 {
		return "我也不知道...", fmt.Errorf("no baidu appid or key config")
	}
	realLen := utils.StringRealLength(str)
	// 当日使用上限
	if checkBaiduOverLimit(realLen) {
		return "超出今日使用字符限制", fmt.Errorf("over upper limit")
	}
	// 语种
	from, to = BaiduCheckLangSupport(from, true), BaiduCheckLangSupport(to, false)
	if len(from) == 0 || len(to) == 0 {
		return "不支持该语种翻译", fmt.Errorf("not support language")
	}
	// 调用API
	ans, err := callBaiduTransAPI(str, from, to)
	// 计数
	if err == nil {
		addBaiduDailyCount(realLen)
	}
	return ans, err
}

var baiduStatus = struct {
	countToday uint64
	countMutex sync.RWMutex
}{}

// 检查所输入的字符串长度是否会超出今日长度限制
func checkBaiduOverLimit(len int) bool {
	baiduStatus.countMutex.RLock()
	defer baiduStatus.countMutex.RUnlock()
	return baiduStatus.countToday+uint64(len) >= uint64(proxy.GetConfigInt64("baidu.maxdaily"))
}

// 将翻译过的字数添加至每日字数统计
func addBaiduDailyCount(len int) {
	baiduStatus.countMutex.Lock()
	defer baiduStatus.countMutex.Unlock()
	baiduStatus.countToday += uint64(len)
}

// 初始化每日字数统计
func initialBaiduDailyCount() {
	// 初始化今日调用字符数 = 0
	baiduStatus.countMutex.Lock()
	baiduStatus.countToday = 0
	baiduStatus.countMutex.Unlock()
}

var baiduTransAPIURL = "https://fanyi-api.baidu.com/api/trans/vip/translate"

type baiduTransResponse struct {
	ErrorCode   int    `json:"error_code"`
	From        string `json:"from"`
	To          string `json:"to"`
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
}

func callBaiduTransAPI(str, from, to string) (string, error) {
	sign, salt := signAsBaiduTransAPI(str)
	data := map[string]string{
		"q":     str,
		"from":  from,
		"to":    to,
		"appid": proxy.GetConfigString("baidu.appid"),
		"salt":  salt,
		"sign":  sign,
	}
	c := client.NewHttpClient(nil)
	rsp, err := c.PostFormByMap(baiduTransAPIURL, data)
	if err != nil {
		return "翻译失败...", fmt.Errorf("baidu post err: %v", err)
	}
	// 解析
	resBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "翻译失败...", fmt.Errorf("baidu read response err: %v", err)
	}
	var res baiduTransResponse
	err = json.Unmarshal(resBytes, &res)
	if err != nil { // 回包解析失败
		return "翻译失败...", fmt.Errorf("baidu unmarshal fail: %v", err)
	}
	if res.ErrorCode != 0 && res.ErrorCode != 52000 { // 含错误码
		return "翻译失败...", fmt.Errorf("baidu error code: %v", res.ErrorCode)
	}
	if len(res.TransResult) == 0 {
		return "我也不知道...", fmt.Errorf("baidu no answer")
	}
	log.Debugf("翻译By Baidu API cost len %v : (%v->%v):%v->%v", utils.StringRealLength(res.TransResult[0].Src),
		res.From, res.To, res.TransResult[0].Src, res.TransResult[0].Dst)
	return res.TransResult[0].Dst, nil
}

func signAsBaiduTransAPI(str string) (sign, salt string) {
	appid := proxy.GetConfigString("baidu.appid")
	key := proxy.GetConfigString("baidu.key")
	salt = strconv.Itoa(rand.Intn(math.MaxInt)) + strconv.Itoa(rand.Intn(math.MaxInt))
	query := appid + str + salt + key
	sign = fmt.Sprintf("%x", md5.Sum([]byte(query)))
	return
}

var BaiduTransLangMap = map[string]string{ // 百度翻译API标准版所支持的语种映射
	"":       "auto",
	"自动":     "auto",
	"自动检测":   "auto",
	"中文":     "zh",
	"汉语":     "zh",
	"英语":     "en",
	"粤语":     "yue",
	"文言文":    "wyw",
	"日语":     "jp",
	"日文":     "jp",
	"韩语":     "kor",
	"韩文":     "kor",
	"法语":     "fra",
	"西语":     "spa",
	"西班牙语":   "spa",
	"泰语":     "th",
	"阿拉伯语":   "ara",
	"俄语":     "ru",
	"葡萄牙语":   "pt",
	"德语":     "de",
	"意大利语":   "it",
	"希腊语":    "el",
	"荷兰语":    "nl",
	"波兰语":    "pl",
	"保加利亚语":  "bul",
	"爱沙尼亚语":  "est",
	"丹麦语":    "dan",
	"芬兰语":    "fin",
	"捷克语":    "cs",
	"罗马尼亚语":  "rom",
	"斯洛文尼亚语": "slo",
	"瑞典语":    "swe",
	"匈牙利语":   "hu",
	"繁体":     "cht",
	"繁体中文":   "cht",
	"越南语":    "vie",
}
