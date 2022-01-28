package music

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "点歌",
	Usage: `点首歌吧
用法：
	点歌 [关键词]+：关键词可以有多个，以空格分隔`,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"music", "点歌"}).SetBlock(true).ThirdPriority().Handle(dealMusic)
}

func dealMusic(ctx *zero.Ctx) {
	args := strings.TrimSpace(utils.GetArgs(ctx))

	if len(args) == 0 { // 未填写歌曲名
		ctx.Send("歌名呢？")
		return
	}

	// 网易云
	id := queryNeteaseMusic(args)
	if id == 0 {
		ctx.Send(fmt.Sprintf("%v没有找到这首歌欸", utils.GetBotNickname()))
	} else {
		ctx.Send(message.Music("163", id))
	}
}

func queryNeteaseMusic(musicName string) int64 {
	c := client.NewHttpClient(nil)
	c.SetUserAgent()
	rsp, err := c.GetGJson("http://music.163.com/api/search/get?type=1&s=" + url.QueryEscape(musicName))
	if err != nil {
		log.Warnf("搜索歌曲[%v]失败：%v", musicName, err)
		return 0
	}
	return rsp.Get("result.songs.0.id").Int()
}
