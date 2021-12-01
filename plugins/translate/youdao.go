package translate

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
)

var youdaoURL = "http://fanyi.youdao.com/translate?smartresult=dict&smartresult=rule&smartresult=ugc&sessionFrom=null"

var youdaoTransLangMap = map[string]string{ // 支持语种映射
	"中文": "zh",
	"英语": "en",
	"日语": "jp",
	"韩语": "kor",
	"汉语": "zh",
	"日文": "jp",
	"韩文": "kor",
}

var youdaoTypeSetMap = map[string]string{ // Type映射
	"zhen":  "ZH_CN2EN",
	"zhjp":  "ZH_CN2JA",
	"zhkor": "ZH_CN2KR",
	"enzh":  "EN2ZH_CN",
	"jpzh":  "JA2ZH_CN",
	"korzh": "KR2ZH_CN",
}

// FreeCheckLangSupport 检查免费翻译是否支持该语种
func FreeCheckLangSupport(lang string) string {
	if trans, ok := youdaoTransLangMap[lang]; ok {
		return trans
	}
	for _, trans := range youdaoTransLangMap {
		if trans == lang {
			return trans
		}
	}
	return ""
}

// FreeTranslate 免费翻译
func FreeTranslate(str, from, to string) (string, error) {
	// 检查
	from, to = FreeCheckLangSupport(from), FreeCheckLangSupport(to)
	if len(from) == 0 || from == "auto" { // 伪自动查找语言
		from = autoFindLang(str)
	}
	if from != "zh" && to != "zh" {
		return "源语言和目的语言至少有一个需要是中文", fmt.Errorf("not support language")
	}
	tp, ok := youdaoTypeSetMap[from+to]
	if !ok {
		return "不支持该语种翻译", fmt.Errorf("not support language")
	}
	// 调用免费 有道API 翻译
	data := map[string]string{
		"type":       tp,
		"i":          str,
		"doctype":    "json",
		"version":    "2.1",
		"keyfrom":    "fanyi.web",
		"ue":         "UTF-8",
		"action":     "FY_BY_CLICKBUTTON",
		"typoResult": "true",
	}
	c := client.NewHttpClient(nil)
	c.SetUserAgent()
	rsp, err := c.PostFormByMap(youdaoURL, data)
	if err != nil {
		return "翻译失败...", err
	}
	// 解析
	res := client.ParseReader(rsp.Body)
	trans := res.Get("translateResult").Array()
	if len(trans) == 0 || len(trans[0].Array()) == 0 {
		return "我也不知道...", nil
	}
	ans := trans[0].Array()[0].Get("tgt").String()
	if len(ans) == 0 {
		return "我也不知道...", nil
	}
	log.Debugf("翻译By Free Youdao : (%v->%v):%v->%v", from, to, str, ans)
	return ans, nil
}

func autoFindLang(str string) string {
	if utils.IsLetter(str) {
		return "en"
	}
	for _, c := range str {
		if 0x3040 <= c && c <= 0x31FF { // 日语
			return "jp"
		}
		if 0xAC00 <= c && c <= 0xD7AF { // 韩语
			return "kor"
		}
		if 0x4E00 <= c && c <= 0x9FD5 { // 中文
			return "zh"
		}
	}
	return "en" // 其它认为是一律英文...
}
