package pixiv_rank

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/plugins/pixiv"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	"github.com/tidwall/gjson"
)

func getPixivRankByHIBI(rankType string, num int, date string) ([]pixiv.PictureInfo, error) {
	// 整理API URL
	api := proxy.GetAPIConfig(consts.APIOfHibiAPIKey)
	if len(api) == 0 {
		return nil, fmt.Errorf("API of HibiAPI is empty")
	}
	if !strings.HasPrefix(api, "http://") && !strings.HasPrefix(api, "https://") {
		api = "https://" + api
	}
	if !strings.HasSuffix(api, "/") {
		api += "/"
	}
	api = fmt.Sprintf("%sapi/pixiv/rank?mode=%v&size=%v", api, rankType, num)
	if len(date) > 0 {
		api += "&date=" + date
	}
	// 调用
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(api)
	if err != nil {
		return nil, err
	}
	rsp = rsp.Get("illusts")
	if !rsp.Exists() {
		return nil, fmt.Errorf("illusts is not found")
	}
	// 解析
	var pics []pixiv.PictureInfo
	illusts := rsp.Array()
	for i, illust := range illusts {
		if i >= num { // 限制数量
			break
		}
		pics = append(pics, analysisHIBIIllust(illust))
	}
	return pics, nil
}

// 分析单张插图信息
func analysisHIBIIllust(illust gjson.Result) pixiv.PictureInfo {
	pic := pixiv.PictureInfo{
		Title:  illust.Get("title").String(),
		PID:    illust.Get("id").Int(),
		P:      0,
		Author: illust.Get("user.name").String(),
		UID:    illust.Get("user.id").Int(),
	}
	for _, tag := range illust.Get("tags").Array() {
		if tag.Get("name").Type != gjson.Null {
			pic.Tags = append(pic.Tags, tag.Get("name").String())
		}
		if tag.Get("translated_name").Type != gjson.Null {
			pic.Tags = append(pic.Tags, tag.Get("translated_name").String())
		}
	}
	pic.Tags = utils.MergeStringSlices(pic.Tags) // 去重
	if illust.Get("page_count").Int() == 1 {
		pic.URL = illust.Get("meta_single_page.original_image_url").String()
	} else if illust.Get("page_count").Int() > int64(pic.P) {
		pic.URL = illust.Get("meta_pages." + strconv.Itoa(pic.P) + ".image_urls.original").String()
	}
	return pic
}
