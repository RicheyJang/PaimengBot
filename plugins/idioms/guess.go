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
	猜不出来的话，发送"算了"或"不知道"结束游戏
	
	猜成语排行榜：按成功猜出成语个数排名的用户总排行榜
	猜成语群排行榜：按成功猜出成语个数排名的本群排行榜`,
	SuperUsage: `
config-plugin文件配置项：
	idioms.localfirst: 是(true)否(false)优先使用本地成语图片，图片放于data/img/idioms目录即可，文件名为答案`,
	Classify: "小游戏",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnFullMatch([]string{"猜成语"}).SetBlock(true).SetPriority(9).Handle(guessIdioms)
	proxy.OnFullMatch([]string{"猜成语排行榜", "猜成语总排行榜"}).SetBlock(true).SetPriority(4).Handle(rankHandler)
	proxy.OnFullMatch([]string{"猜成语群排行榜"}, zero.OnlyGroup).SetBlock(true).SetPriority(4).Handle(groupRankHandler)
	proxy.AddConfig("localFirst", false) // 优先使用本地词库IdiomsImageDir, 文件名：某个成语.png/jpg
	_, _ = utils.MakeDir(consts.IdiomsImageDir)
}

var cancelMessage = []string{"算啦", "算了", "cancel", "取消", "不知道"}

func guessIdioms(ctx *zero.Ctx) {
	if ctx.Event.GroupID != 0 { // 同一个群，只允许有一个猜成语
		if proxy.LockUser(ctx.Event.GroupID) {
			ctx.Send("群里还有正在猜的成语，先把它猜出来吧")
			return
		}
		defer proxy.UnlockUser(ctx.Event.GroupID)
	}
	// 获取成语图片
	msg, key, err := getIdiomsPicture()
	if err != nil {
		log.Errorf("getIdiomsPicture err: %v", err)
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		return
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text(`猜不出来的话，跟我说"算了"或者"不知道"`), msg)
	log.Infof("正确答案：%v", key)
	// 等待用户回复
	r, cancel := ctx.FutureEvent("message", func(futureCtx *zero.Ctx) bool {
		if futureCtx.Event.GroupID == 0 { // 私聊消息
			return ctx.Event.GroupID == 0 && ctx.Event.UserID == futureCtx.Event.UserID
		} else if ctx.Event.GroupID == futureCtx.Event.GroupID { // 同一个群的群消息
			guess := futureCtx.Event.Message.ExtractPlainText()
			if ((futureCtx.Event.UserID == ctx.Event.UserID || utils.IsGroupAdmin(futureCtx)) && utils.StringSliceContain(cancelMessage, guess)) ||
				guess == key { // 发起人或群管取消 或 有人猜对了答案，处理
				return true
			}
		}
		return false
	}).Repeat()
	defer cancel()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	// 事件循环
	for {
		select {
		case e := <-r:
			if e == nil {
				return
			}
			guess := strings.TrimSpace(e.Message.ExtractPlainText())
			if guess == key { // 猜对，结束游戏
				ctx.SendChain(message.At(e.UserID), message.Text("猜对啦"))
				addSuccess(e.GroupID, e.UserID)
				return
			} else if utils.StringSliceContain(cancelMessage, guess) { // 取消，结束游戏
				ctx.SendChain(message.At(e.UserID), message.Text(fmt.Sprintf("那算啦，其实正确答案是%v哦", key)))
				return
			} else { // 猜错，继续游戏
				if ctx.Event.GroupID == 0 { // 只有私聊提示
					ctx.Send(message.Text("猜错了哦"))
				}
			}
		case <-ticker.C: // 超时取消
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("太久啦，其实正确答案是%v哦", key)))
			return
		}
	}
}

// 获取猜成语图片：

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
