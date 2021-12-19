package idioms

import (
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "猜成语",
	Usage: `用法：
	猜成语：扔给你一张图片，猜出来是什么成语吧
	猜不出来的话，发送"算了"或"不知道"结束游戏`,
	Classify: "小游戏",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"猜成语"}).SetBlock(true).ThirdPriority().Handle(guessIdioms)
	proxy.AddConfig("localFirst", false) // 优先使用本地词库IdiomsImageDir, 文件名：某个成语.png/jpg
	_, _ = utils.MakeDir(consts.IdiomsImageDir)
}

var cancelMessage = []string{"算啦", "算了", "cancel", "取消", "不知道"}

func guessIdioms(ctx *zero.Ctx) {
	msg, key, err := getIdiomsPicture()
	if err != nil {
		log.Errorf("getIdiomsPicture err: %v", err)
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		return
	}
	ctx.SendChain(message.At(ctx.Event.UserID), msg)
	log.Infof("正确答案：%v", key)
	// 等待用户回复
	r, cancel := ctx.FutureEvent("message", ctx.CheckSession()).Repeat()
loop:
	for {
		select {
		case e := <-r:
			guess := strings.TrimSpace(e.Message.ExtractPlainText())
			if utils.StringSliceContain(cancelMessage, guess) { // 取消，结束游戏
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("那算啦，其实正确答案是%v哦", key)))
				cancel()
				break loop
			} else if guess == key { // 猜对，结束游戏
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("猜对啦"))
				cancel()
				break loop
			} else { // 猜错，继续游戏
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("猜错了哦"))
			}
		case <-time.After(5 * time.Minute): // 超时取消
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("太久啦，其实正确答案是%v哦", key)))
			cancel()
			break loop
		}
	}
}

func getIdiomsPicture() (msg message.MessageSegment, key string, err error) {
	if proxy.GetConfigBool("localFirst") {
		msg, key, err = getIdiomsPictureLocal()
		if err == nil {
			return
		}
	}
	// 尝试API
	msg, key, err = getIdiomsPictureByIYK0()
	if err != nil {
		return getIdiomsPictureLocal()
	}
	return
}

func getIdiomsPictureLocal() (msg message.MessageSegment, key string, err error) {
	// 计数
	count := 0
	_ = filepath.WalkDir(consts.IdiomsImageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		ext := filepath.Ext(d.Name())
		if d.IsDir() || len(d.Name()) <= 4 || len(ext) == 0 || !(ext == ".jpg" || ext == ".png") {
			return nil
		}
		count += 1
		return nil
	})
	if count == 0 {
		return message.MessageSegment{}, "", fmt.Errorf("%v is empty", consts.IdiomsImageDir)
	}
	// 随机选取
	num := rand.Int() % count
	err = filepath.WalkDir(consts.IdiomsImageDir, func(path string, d fs.DirEntry, ferr error) error {
		if ferr != nil {
			return ferr
		}
		ext := filepath.Ext(d.Name())
		if d.IsDir() || len(d.Name()) <= 4 || len(ext) == 0 || !(ext == ".jpg" || ext == ".png") {
			return nil
		}
		count -= 1
		if count == num {
			msg, err = utils.GetImageFileMsg(path)
			key = d.Name()[:len(d.Name())-4]
			return err
		}
		return nil
	})
	if err != nil {
		return message.MessageSegment{}, "", fmt.Errorf("filepath walk err: %v", err)
	}
	if len(key) == 0 {
		return message.MessageSegment{}, "", fmt.Errorf("key is empty")
	}
	return
}
