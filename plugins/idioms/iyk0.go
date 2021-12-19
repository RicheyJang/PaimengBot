package idioms

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const idiomsPictureAPI = "https://api.iyk0.com/ktc/"

func getIdiomsPictureByIYK0() (msg message.MessageSegment, key string, err error) {
	c := client.NewHttpClient(&client.HttpOptions{TryTime: 2})
	rsp, err := c.GetGJson(idiomsPictureAPI)
	if err != nil {
		return message.MessageSegment{}, "", err
	}
	if rsp.Get("code").Int() != 200 {
		return message.MessageSegment{}, "", fmt.Errorf("rsp code != 200, msg: %v", rsp.Get("msg"))
	}
	key = rsp.Get("key").String()
	url := rsp.Get("img").String()
	go saveIdiomsPictureByIYK0ToLocal(url, key)
	return message.Image(url), key, nil
}

func saveIdiomsPictureByIYK0ToLocal(url string, key string) {
	dot := strings.LastIndex(url, ".")
	if dot <= 0 {
		return
	}
	ext := url[dot:]
	path := utils.PathJoin(consts.IdiomsImageDir, "iyk0")
	_, _ = utils.MakeDir(path)
	path = utils.PathJoin(path, key+ext)
	err := client.DownloadToFile(path, url, 3)
	if err == nil {
		log.Infof("success to download idioms picture %v to %v", key, path)
	} else {
		log.Infof("failed to download, err: %v", err)
	}
}
