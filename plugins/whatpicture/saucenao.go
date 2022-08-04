package whatpicture

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func searchPicBySaucenao(picURL string, showAdult bool) ([]message.Message, error) {
	numres := proxy.GetConfigInt64("max")
	if numres <= 0 {
		numres = 1
	}
	// 调用API
	api := fmt.Sprintf("https://saucenao.com/search.php?db=999&output_type=2&numres=%d&url=%s&api_key=%s",
		numres, url.QueryEscape(picURL), proxy.GetConfigString("key"))
	if !showAdult {
		api += "&hide=2"
	}
	timeoutS := proxy.GetConfigString("timeout")
	timeout, err := time.ParseDuration(timeoutS)
	if err != nil || timeout <= 0 {
		timeout = 30 * time.Second
	}
	cli := client.NewHttpClient(&client.HttpOptions{Timeout: timeout})
	rsp, err := cli.GetGJson(api)
	if err != nil {
		return []message.Message{{message.Text("出错了...")}}, err
	}
	// 解析
	var msgs []message.Message
	minSimilarity := rsp.Get("header.minimum_similarity").Float()
	for _, result := range rsp.Get("results").Array() {
		if result.Get("header.similarity").Float() < minSimilarity {
			continue
		}
		picInfo := newPictureInfo(result)
		// 分Index讨论分别解析
		switch result.Get("header.index_id").Int() {
		case 5, 6, 51, 52, 53: // pixiv
			picInfo.parsePixivInfo(result)
		case 18, 38: // nhentai, e-hentai
			picInfo.parseHMiscInfo(result)
		case 21, 22: // Anime
			picInfo.parseAnimeInfo(result)
		}
		// 默认解析
		if len(picInfo.exDescribe) == 0 {
			picInfo.parseDefault(result)
		}
		msgs = append(msgs, picInfo.genMessage())
	}
	if len(msgs) == 0 {
		return []message.Message{{message.Text("没有找到相似图片")}}, nil
	}
	return msgs, nil
}

type pictureInfo struct {
	// 通用
	thumbnail  string
	similarity float64
	srcURL     string
	srcDB      string
	hidden     bool
	// 需各解析函数填写
	category   string
	title      string
	exDescribe string
}

var pictureIndexReg = regexp.MustCompile(`^Index\s+#\d+:\s+(.+)\s-\s.+`)

func newPictureInfo(result gjson.Result) *pictureInfo {
	p := &pictureInfo{
		thumbnail:  result.Get("header.thumbnail").String(),
		similarity: result.Get("header.similarity").Float(),
		hidden:     result.Get("header.hidden").Bool(),
		srcURL:     result.Get("data.ext_urls.0").String(),
	}
	sub := pictureIndexReg.FindStringSubmatch(result.Get("header.index_name").String())
	if len(sub) > 1 {
		p.srcDB = sub[1]
	}
	return p
}

// 各类数据解析器

// 解析Pixiv
func (p *pictureInfo) parsePixivInfo(result gjson.Result) {
	p.category = "插画"
	p.title = result.Get("data.title").String()
	var str string
	if result.Get("data.pixiv_id").Exists() {
		str = fmt.Sprintf("PID: %d\n", result.Get("data.pixiv_id").Int())
	}
	if result.Get("data.member_name").Exists() {
		str += fmt.Sprintf("作者: %s\n", result.Get("data.member_name").String())
	}
	if result.Get("data.member_id").Exists() {
		str += fmt.Sprintf("UID: %d\n", result.Get("data.member_id").Int())
	}
	p.exDescribe = strings.TrimSpace(str)
}

// 解析HMisc
func (p *pictureInfo) parseHMiscInfo(result gjson.Result) {
	p.category = "本子"
	if result.Get("data.jp_name").String() != "" {
		p.title = result.Get("data.jp_name").String()
	} else if result.Get("data.eng_name").String() != "" {
		p.title = result.Get("data.eng_name").String()
	} else if result.Get("data.source").String() != "" {
		p.title = result.Get("data.source").String()
	}
	// 描述 仅有作者
	var creator []string
	for i, c := range result.Get("data.creator").Array() {
		if i >= 5 { // 最多展示5个
			creator = append(creator, "...")
			break
		}
		creator = append(creator, c.String())
	}
	if len(creator) > 0 {
		p.exDescribe = fmt.Sprintf("作者: %s", strings.Join(creator, "、"))
	}
}

// 解析动漫
func (p *pictureInfo) parseAnimeInfo(result gjson.Result) {
	p.category = "动漫"
	p.title = result.Get("data.source").String()
	p.exDescribe = fmt.Sprintf("第%s集%s",
		result.Get("data.part").String(), result.Get("data.est_time").String())
}

// saucenao返回数据的默认解析方法
func (p *pictureInfo) parseDefault(result gjson.Result) {
	if len(p.title) == 0 { // 尝试标题
		if result.Get("data.title").String() != "" {
			p.title = result.Get("data.title").String()
		} else if result.Get("data.jp_name").String() != "" {
			p.title = result.Get("data.jp_name").String()
		} else if result.Get("data.eng_name").String() != "" {
			p.title = result.Get("data.eng_name").String()
		} else if result.Get("data.source").String() != "" {
			p.title = result.Get("data.source").String()
		} else if result.Get("data.*_name").Exists() {
			p.title = result.Get("data.*_name").String()
		}
	}
	// 补全描述
	if result.Get("data.part").Exists() {
		p.exDescribe += "\n集数：" + result.Get("data.part").String()
	}
	if result.Get("data.est_time").Exists() {
		p.exDescribe += "\n时间：" + result.Get("data.est_time").String()
	}
	if result.Get("data.author").Exists() {
		p.exDescribe += "\n作者：" + result.Get("data.author").String()
	} else if result.Get("data.author_name").Exists() {
		p.exDescribe += "\n作者：" + result.Get("data.author_name").String()
	}
	if result.Get("data.company").Exists() {
		p.exDescribe += "\n厂商：" + result.Get("data.company").String()
	}
	p.exDescribe = strings.TrimSpace(p.exDescribe)
}

// 生成消息
func (p pictureInfo) genMessage() (msg message.Message) {
	// 标题
	title := p.category
	if len(title) != 0 {
		title += ": "
	}
	if len(p.title) != 0 {
		title += p.title
	}
	if len(title) != 0 {
		msg = append(msg, message.Text(title))
	}
	// 缩略图
	if !p.hidden && len(p.thumbnail) != 0 {
		msg = append(msg, message.Image(p.thumbnail))
	}
	// 描述
	str := fmt.Sprintf("相似度: %.2f%%\n来源: %s", p.similarity, p.srcDB)
	if len(p.exDescribe) != 0 {
		str += "\n" + p.exDescribe
	}
	if proxy.GetConfigBool("link") && len(p.srcURL) > 0 {
		str += "\n链接: " + p.srcURL
	}
	msg = append(msg, message.Text(str))
	return
}
