package whatanime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func searchAnimeByTraceMoe(url string, showAdult bool) (msg message.Message, err error) {
	api := proxy.GetAPIConfig(consts.APIOfTraceMoeKey)
	if len(api) == 0 {
		return message.Message{message.Text("失败了...")}, fmt.Errorf("api of trace.moe is empty")
	}
	// 整理API URL
	if !strings.HasPrefix(api, "http://") && !strings.HasPrefix(api, "https://") {
		api = "https://" + api
	}
	if !strings.HasSuffix(api, "/") {
		api += "/"
	}
	// 暂时使用trace.moe提供的anilistInfo，今后可以改成https://anilist.gitbook.io/anilist-apiv2-docs/提供的更为详细的信息
	api = fmt.Sprintf("%ssearch?anilistInfo&cutBorders&url=%s", api, url)
	// 调用
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(api)
	if err != nil {
		return message.Message{message.Text("出错了...")}, err
	}
	results := rsp.Get("result").Array()
	if len(results) == 0 { // 没有结果
		return message.Message{message.Text(fmt.Sprintf("%v也不知道", utils.GetBotNickname()))},
			fmt.Errorf("result is empty, error=%v", rsp.Get("error"))
	}
	result := results[0]
	// 解析
	isAdult := result.Get("anilist").Get("isAdult").Bool()
	title := formatMoeResultTitle(result)
	imgMsg := message.Image(result.Get("image").String()) // 直接以URL格式发送
	if isAdult && !showAdult {
		imgMsg = message.Text("\n不给你看图\n")
	}
	text := fmt.Sprintf("相似度：%v\n位置：第%v集的%v", formatMoeResultSimilarity(result.Get("similarity")),
		getMoeResultEpisode(result.Get("episode")), formatMoeResultTime(result.Get("from")))
	return message.Message{message.Text(title), imgMsg, message.Text(text)}, nil
}

func formatMoeResultSimilarity(similarity gjson.Result) string {
	if !similarity.Exists() {
		return "未知"
	}
	org := similarity.Float() * 100
	str := strconv.FormatFloat(org, 'f', 2, 64) + "%"
	if org <= 90 {
		str += "(较低)"
	}
	return str
}

func getMoeResultEpisode(episode gjson.Result) string {
	if !episode.Exists() {
		return "?"
	}
	i := episode.Int()
	return strconv.FormatInt(i, 10)
}

func formatMoeResultTime(from gjson.Result) string {
	if !from.Exists() {
		return "未知时间"
	}
	org := math.Floor(from.Float())
	return fmt.Sprintf("%02d:%02d", int(org)/60, int(org)%60)
}

func formatMoeResultTitle(result gjson.Result) string {
	title := result.Get("anilist").Get("title")
	if !title.Exists() {
		return result.Get("filename").String()
	}
	res := title.Get("native").String()
	if title.Get("chinese").Type != gjson.Null {
		res += "\n" + title.Get("chinese").String()
	} else if title.Get("english").Type != gjson.Null {
		res += "\n" + title.Get("english").String()
	} else if title.Get("romaji").Type != gjson.Null {
		res += "\n" + title.Get("romaji").String()
	}
	return res
}
