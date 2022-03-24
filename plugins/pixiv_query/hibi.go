package pixiv_query

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

func getPixivPIDsByHIBI(pid int64) ([]pixiv.PictureInfo, error) {
	// URL
	api, err := getHibiAPI()
	if err != nil {
		return nil, err
	}
	api = fmt.Sprintf("%sapi/pixiv/illust?id=%v", api, pid)
	// 调用
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(api)
	if err != nil {
		return nil, err
	}
	illust := rsp.Get("illust")
	if !illust.Exists() {
		return []pixiv.PictureInfo{{Title: rsp.Get("error.user_message").String()}}, fmt.Errorf("illust is not found")
	}
	// 解析
	pics := make([]pixiv.PictureInfo, int(illust.Get("page_count").Int()))
	if len(pics) == 0 {
		return nil, fmt.Errorf("page_count is zero")
	}
	pics[0] = analysisHIBIIllust(illust, 0)
	if len(pics) > 1 { // 获取所有分P的URL
		urls := illust.Get("meta_pages.#.image_urls.original").Array()
		for i, url := range urls {
			if i >= len(pics) {
				break
			}
			pics[i].PID = pid
			pics[i].P = i
			pics[i].URL = url.String()
		}
	}
	return pics, nil
}

// 工具函数

// 整理API URL
func getHibiAPI() (string, error) {
	api := proxy.GetAPIConfig(consts.APIOfHibiAPIKey)
	if len(api) == 0 {
		return "", fmt.Errorf("API of HibiAPI is empty")
	}
	if !strings.HasPrefix(api, "http://") && !strings.HasPrefix(api, "https://") {
		api = "https://" + api
	}
	if !strings.HasSuffix(api, "/") {
		api += "/"
	}
	return api, nil
}

// 分析单张插图信息
func analysisHIBIIllust(illust gjson.Result, p int) pixiv.PictureInfo {
	pic := pixiv.PictureInfo{
		Title:  illust.Get("title").String(),
		PID:    illust.Get("id").Int(),
		P:      p,
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
