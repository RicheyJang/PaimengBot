package pixiv

import (
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
)

// 从Lolicon API随机获取图片，返回图片URL切片
func getPicturesFromLolicon(tags []string, num int, isR18 bool) []PictureInfo {
	if num <= 0 {
		log.Warnf("num <= 0, num: %d", num)
		return []PictureInfo{}
	}
	// 请求
	c := client.NewHttpClient(&client.HttpOptions{TryTime: 3, Timeout: 15 * time.Second})
	req := APIRequest{
		Num:   num,
		Tags:  tags,
		Size:  []string{"original"},
		Proxy: proxy.GetConfigString("proxy"),
		R18:   0,
	}
	if isR18 {
		req.R18 = 1
	}
	rsp := APIResponse{}
	err := c.PostMarshal(PixivAPI, req, &rsp)
	if err != nil {
		log.Warnf("PostMarshal %v, err: %v", PixivAPI, err)
		return []PictureInfo{}
	}
	// 解析
	if len(rsp.Data) == 0 {
		log.Warnf("PostMarshal %v rsp is empty, req=%+v, rsp.error=%v", PixivAPI, req, rsp.Error)
		return []PictureInfo{}
	}
	var pics []PictureInfo
	for _, data := range rsp.Data {
		if !isR18 && (data.R18 || utils.StringSliceContain(data.Tags, "R-18")) { // 该图片包含R-18标签，也跳过
			continue
		}
		pic := PictureInfo{
			Title:  data.Title,
			URL:    data.Urls["original"],
			Tags:   data.Tags,
			PID:    data.Pid,
			P:      data.P,
			Author: data.Author,
			UID:    data.Uid,
		}
		pics = append(pics, pic)
	}
	return pics
}

const PixivAPI = "https://api.lolicon.app/setu/v2"

type APIRequest struct {
	Num   int      `json:"num,omitempty"`
	Tags  []string `json:"tag,omitempty"`
	Size  []string `json:"size,omitempty"`
	Proxy string   `json:"proxy,omitempty"`
	R18   int      `json:"r18"`
}

type APIResponse struct {
	Error string          `json:"error"`
	Data  []SinglePicture `json:"data"`
}

type SinglePicture struct {
	Pid        int64             `json:"pid"`
	P          int               `json:"p"`
	Uid        int64             `json:"uid"`
	Title      string            `json:"title"`
	Author     string            `json:"author"`
	R18        bool              `json:"r18"`
	Width      int               `json:"width"`
	Height     int               `json:"height"`
	Tags       []string          `json:"tags"`
	Ext        string            `json:"ext"`
	UploadDate int               `json:"uploadDate"`
	Urls       map[string]string `json:"urls"`
}
