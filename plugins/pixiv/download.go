package pixiv

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type downloader struct {
	has  bool
	pics []PictureInfo
	cap  int

	tags  []string
	num   int
	isR18 bool
}

func newDownloader(tags []string, num int, isR18 bool) *downloader {
	var realTags []string
	for _, tag := range tags {
		if len(tag) > 0 {
			realTags = append(realTags, tag)
		}
	}
	return &downloader{
		has: false,
		cap: (num + 5) * 2, // 为防止后续图片下载失败等，拿取的图片信息数会>num

		tags:  realTags,
		num:   num,
		isR18: isR18,
	}
}

func (d *downloader) get() {
	d.has = true
	// 从getter获取图片信息
	sum := 0
	for k, _ := range getterMap {
		sum += int(proxy.GetConfigInt64(fmt.Sprintf("scale.%s", k)))
	}
	for k, getter := range getterMap {
		scale := proxy.GetConfigInt64(fmt.Sprintf("scale.%s", k))
		if scale == 0 { // 设定比例为0，跳过
			continue
		}
		single := float64(scale) / float64(sum)
		pics := getter(d.tags, int(float64(d.cap)*single)+1, d.isR18)
		for i := range pics { // 标注来源图库
			pics[i].Src = k
		}
		d.pics = append(d.pics, pics...)
	}
	// 重排序
	rand.Shuffle(len(d.pics), func(i int, j int) {
		d.pics[i], d.pics[j] = d.pics[j], d.pics[i]
	})
	sort.Slice(d.pics, func(i, j int) bool { // 优先已有URL的
		return len(d.pics[i].URL) > len(d.pics[j].URL)
	})
}

func (d *downloader) send(ctx *zero.Ctx) {
	if !d.has {
		d.get()
	}
	// 预处理图片信息
	if len(d.pics) == 0 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("没图了..."))
		return
	}
	// 下载图片
	var i, num int
	for i, num = 0, 0; i < len(d.pics) && num < d.num; i++ {
		msg, err := d.pics[i].GenSinglePicMsg() // 生成图片消息
		if err == nil {                         // 成功
			ctx.Send(msg)
			log.Infof("发送Pixiv图片成功 pid=%v, 来源：%v", d.pics[i].PID, d.pics[i].Src)
			num += 1
		} else { // 失败
			log.Infof("生成Pixiv消息失败 url=%v, 来源=%v, err=%v", d.pics[i].URL, d.pics[i].Src, err)
		}
	}
	if num == 0 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
	}
}

// GenSinglePicMsg 生成单条Pixiv消息
func (pic *PictureInfo) GenSinglePicMsg() (message.Message, error) {
	// 初始化
	if pic == nil {
		return nil, fmt.Errorf("pic is nil")
	}
	if len(pic.URL) == 0 {
		err := pic.GetURLByPID()
		if err != nil {
			return nil, fmt.Errorf("GetURLByPID failed: %v", err)
		}
	}
	// 下载图片
	path, err := images.GetNewTempSavePath("pixiv")
	if err != nil {
		return nil, err
	}
	c := client.NewHttpClient(&client.HttpOptions{TryTime: 2, Timeout: getTimeout()})
	err = c.DownloadToFile(path, pic.URL)
	if err != nil {
		return nil, err
	}
	// 构成消息
	picMsg, err := utils.GetImageFileMsg(path)
	if err != nil {
		return nil, err
	}
	// 文字
	tip := pic.GetDescribe()
	return message.Message{message.Text(pic.Title), picMsg, message.Text(tip)}, nil
}

// GetDescribe 获取图片说明
func (pic *PictureInfo) GetDescribe() string {
	var tags []string
	for _, tag := range pic.Tags {
		if isCNOrEn(tag) {
			tags = append(tags, tag)
		}
	}
	tip := fmt.Sprintf("PID: %v", pic.PID)
	if pic.P != 0 {
		tip += fmt.Sprintf("(p%d)", pic.P)
	}
	if len(pic.Author) > 0 {
		tip += fmt.Sprintf("\n作者: %v", pic.Author)
	}
	if pic.UID != 0 {
		tip += fmt.Sprintf("\nUID: %v", pic.UID)
	}
	if len(tags) > 0 {
		tip += fmt.Sprintf("\n标签: %v", strings.Join(tags, ","))
	}
	return tip
}

// GetURLByPID 通过PID获取图片下载URL
func (pic *PictureInfo) GetURLByPID() (err error) {
	if pic.PID == 0 {
		return fmt.Errorf("pid is 0")
	}
	// 整理API URL
	api := proxy.GetAPIConfig(consts.APIOfHibiAPIKey)
	if len(api) == 0 {
		return fmt.Errorf("API of HibiAPI is empty")
	}
	if !strings.HasPrefix(api, "http://") && !strings.HasPrefix(api, "https://") {
		api = "https://" + api
	}
	if !strings.HasSuffix(api, "/") {
		api += "/"
	}
	api = fmt.Sprintf("%sapi/pixiv/illust?id=%v", api, pic.PID)
	// 调用
	c := client.NewHttpClient(nil)
	rsp, err := c.GetGJson(api)
	if err != nil {
		return err
	}
	rsp = rsp.Get("illust")
	if !rsp.Exists() {
		return fmt.Errorf("illust is not found")
	}
	defer func() { // 替换代理
		if err == nil && len(pic.URL) == 0 {
			err = fmt.Errorf("unexpected error")
		} else if len(pic.URL) > 0 {
			pic.ReplaceURLToProxy()
		}
	}()
	// 解析
	if rsp.Get("page_count").Int() == 1 {
		pic.URL = rsp.Get("meta_single_page.original_image_url").String()
	} else if rsp.Get("page_count").Int() > int64(pic.P) {
		pic.URL = rsp.Get("meta_pages." + strconv.Itoa(pic.P) + ".image_urls.original").String()
	}
	return nil
}

// ReplaceURLToProxy 将图片URL替换为反代地址
func (pic *PictureInfo) ReplaceURLToProxy() {
	if len(pic.URL) > 0 {
		p := proxy.GetConfigString("proxy")
		if len(p) > 0 {
			pic.URL = strings.ReplaceAll(pic.URL, "i.pximg.net", p)
			pic.URL = strings.ReplaceAll(pic.URL, "i.pixiv.net", p)
		}
	}
}

func getTimeout() time.Duration {
	s := proxy.GetConfigString("timeout")
	if len(s) == 0 { // 未设置
		return 10 * time.Second
	}
	t, err := time.ParseDuration(s)
	if err != nil || t < time.Second { // 设置错误
		return 10 * time.Second
	}
	return t
}

var asciiReg = regexp.MustCompile(`^[A-Za-z0-9_+,=~!@#<>\[\]{}:/.?'"$%&*()\-\\]+$`)

func isCNOrEn(str string) bool {
	for _, c := range str {
		if 0x3040 <= c && c <= 0x31FF { // 日语
			return false
		}
		if 0xAC00 <= c && c <= 0xD7AF { // 韩语
			return false
		}
	}
	for _, c := range str {
		if 0x4E00 <= c && c <= 0x9FD5 { // 有中文
			return true
		}
	}
	if asciiReg.MatchString(str) {
		return true
	}
	return false
}
