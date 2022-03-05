package genshin_resource

import (
	"fmt"
	"os"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var info = manager.PluginInfo{
	Name: "今日素材",
	Usage: `看看今天都可以打什么材料吧！
用法：
	今日素材：丢给你今天要打(上班)的材料清单`,
	SuperUsage: `
	更新今日素材：强制更新今日素材图片
另，每天4点半会自动更新当天素材`,
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

const imageFilePrefix = "resource"

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"今日素材", "原神今日素材"}).SetBlock(true).SetPriority(3).Handle(sendTodayResource)
	proxy.OnCommands([]string{"更新今日素材"}, zero.SuperUserPermission).SetBlock(true).SetPriority(3).Handle(flushTodayResource)
	_, _ = proxy.AddScheduleDailyFunc(4, 30, func() { _, _ = getTodayResource() }) // 每天四点半尝试更新
}

func sendTodayResource(ctx *zero.Ctx) {
	file := checkTodayResourceFile()
	if len(file) == 0 {
		file, _ = getTodayResource()
	}
	if len(file) == 0 {
		ctx.Send("失败了...")
		return
	}
	msg, err := utils.GetImageFileMsg(file)
	if err != nil {
		ctx.Send("失败了...")
		return
	}
	ctx.Send(msg)
}

func flushTodayResource(ctx *zero.Ctx) {
	if _, err := getTodayResource(); err != nil {
		ctx.Send("失败了...")
	} else {
		ctx.Send("更新成功")
	}
}

// 获取今日素材图片文件名
func getTodayResourceFilename() (string, error) {
	dir, err := utils.MakeDir(utils.PathJoin(consts.GenshinImageDir, "resource"))
	if err != nil {
		log.Warnf("getTodayResourceFilename MakeDir %v err: %v", dir, err)
		return "", err
	}
	filename := utils.PathJoin(dir, fmt.Sprintf("%s-%d.png", imageFilePrefix, time.Now().YearDay()))
	return filename, nil
}

// 检查今日素材图片文件是否存在且有效，有效则返回绝对路径，无效则返回空字符串
func checkTodayResourceFile() string {
	filename, err := getTodayResourceFilename()
	if err != nil || !utils.FileExists(filename) {
		return ""
	}
	fs, _ := os.Stat(filename)
	if fs.Size() == 0 || time.Now().Sub(fs.ModTime()) >= 48*time.Hour {
		return ""
	}
	return filename
}

// 有新截图加入时，修改此列表即可，filename为所需保存至的文件名
var snapshotFunctions = []func(filename string) error{
	getTodayResourceByGenshinPub, // 0 主图 通过genshin.pub获取今日素材图
	getTodayEventByMhyObc,        // 1 通过米游社的观测枢获取今日活动进展图
}

// 获取今日素材图片文件（会强制替换已有文件）
func getTodayResource() (filename string, err error) {
	filename, err = getTodayResourceFilename()
	if err != nil {
		return "", err
	}
	// 截图
	var srcFilenames []string // 成功截图所保存的文件名列表
	for i, f := range snapshotFunctions {
		src := filename // 当前截图保存的文件名
		if i != 0 {     // 非主图（第一个）
			src = fmt.Sprintf("%s-%d.png", filename, i)
		}
		err = f(src) // 截图去
		if err != nil {
			log.Errorf("snapshot func[%d] fail, err=%v", i, err)
			if i == 0 { // 主图失败，直接返回
				return "", err
			}
		} else { // 截图成功，保存至srcFilenames
			srcFilenames = append(srcFilenames, src)
		}
	}
	defer func() {
		// 删除多余文件
		for _, src := range srcFilenames {
			if src != filename {
				_ = os.Remove(src)
			}
		}
	}()
	// 合并文件
	if len(srcFilenames) == 0 {
		return "", fmt.Errorf("no snapshot success")
	} else { // 有至少一个截图成功了
		err = images.MergeImageFile("#f6f2ee", filename, srcFilenames...)
		if err != nil {
			log.Errorf("MergeImageFile err: %v", err)
			return filename, err
		}
	}
	log.Infof("成功更新原神今日素材+活动图片：%v", filename)
	return filename, err
}
