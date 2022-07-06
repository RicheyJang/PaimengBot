package pixiv_query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "pixiv搜索",
	Usage: `用法：(以下功能仅限私聊)
	pixiv搜索 p[PID]：发送指定PID的所有分P插画`,
	SuperUsage: `config-plugin配置项:
	api.hibiapi: 所选用HibiAPI的网址，若失效，可以自行搭建
搭建方法：https://github.com/mixmoe/HibiAPI`,
	Classify: "好康的",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"pixiv搜索", "pixiv查询", "p站搜索", "P站查询"}, zero.OnlyPrivate).SetBlock(true).SetPriority(3).Handle(queryHandler)
	proxy.AddAPIConfig(consts.APIOfHibiAPIKey, "api.obfs.dev")
}

func queryHandler(ctx *zero.Ctx) {
	if proxy.LockUser(ctx.Event.UserID) {
		ctx.Send("有正在进行中的搜索，稍等哦")
		return
	}
	defer proxy.UnlockUser(ctx.Event.UserID)
	// 解析参数
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	if strings.HasPrefix(arg, "p") { // 搜索PID
		id, err := strconv.ParseInt(arg[len("p"):], 10, 64)
		if err != nil {
			ctx.Send("PID格式不对")
			return
		}
		queryPID(ctx, id)
		return
	} else { // 搜索其它内容
		ctx.Send(`暂不支持搜索此类内容，可以看看"帮助 pixiv搜索"或"帮助 好康的"`)
		return
	}
}

func queryPID(ctx *zero.Ctx, pid int64) {
	pics, err := getPixivPIDsByHIBI(pid)
	if err != nil || len(pics) == 0 {
		if len(pics) > 0 && len(pics[0].Title) > 0 {
			ctx.Send(pics[0].Title)
		} else {
			ctx.Send("失败了...")
		}
		log.Warnf("getPixivPIDsByHIBI err: %v", err)
		return
	}
	// 首图信息及模式检查
	mainPic := pics[0]
	str := mainPic.Title + "\n" + mainPic.GetDescribe()
	if !mainPic.CheckNoSESE() {
		ctx.Send(str)
		return
	}
	str += fmt.Sprintf("\n是否确定发送？")
	ctx.Send(str)
	// 发送确认
	if !isConfirm(ctx) {
		ctx.Send("那算啦")
		return
	}
	ctx.Send(fmt.Sprintf("即将发送共%d张图片", len(pics)))
	// 发送图片
	for i, pic := range pics {
		path, err := pic.GetPicture()
		if err != nil { // 下载失败
			log.Infof("下载第%d张插画失败，err: %v", i, err)
			ctx.Send(fmt.Sprintf("下载第%d张插画失败，跳过", i+1))
			continue
		}
		// 构成消息
		picMsg, err := utils.GetImageFileMsg(path)
		if err != nil {
			log.Infof("发送第%d张插画失败 GetImageFileMsg err: %v", i, err)
			ctx.Send(fmt.Sprintf("发送第%d张插画失败，跳过", i+1))
			continue
		}
		ctx.Send(message.Message{picMsg})
	}
}

func isConfirm(ctx *zero.Ctx) bool {
	e := utils.WaitNextMessage(ctx)
	if e == nil {
		return false
	}
	confirm := e.Message.ExtractPlainText()
	return confirm == "是" || confirm == "确定" || confirm == "确认"
}
