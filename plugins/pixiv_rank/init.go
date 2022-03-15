package pixiv_rank

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "pixiv排行榜",
	Usage: `用法：
	pixiv排行榜 [类型]? [数量]? [日期]?：获取指定日期指定数量指定类型的Pixiv排行榜图片
可选类型（左侧数字）：
	1. 日排行
	2. 周排行
	3. 月排行
	4. 男性向日排行
	5. 女性向日排行
	6. 原创排行
	7. 新人排行
以下类型仅限私聊：
	8. R18日排行
	9. R18周排行
	10. R18男性向日排行
	11. R18女性向日排行
示例：
	pixiv排行榜：默认当日排行5张
	pixiv排行榜 4：男性向当日排行5张
	pixiv排行榜 4 20：男性向当日排行20张
	pixiv排行榜 4 20 2022-01-28：男性向2022年1月28日排行10张`,
	SuperUsage: `
备注：config-plugin配置项沿用pixiv插件`,
	Classify: "好康的",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"pixiv排行榜", "Pixiv排行榜"}).SetBlock(true).SetPriority(3).Handle(getRankPictures)

	proxy.AddAPIConfig(consts.APIOfHibiAPIKey, "api.obfs.dev")
}

func getRankPictures(ctx *zero.Ctx) {
	if proxy.LockUser(ctx.Event.UserID) {
		ctx.Send("正在发送中呢，稍等一下嘛")
		return
	}
	defer proxy.UnlockUser(ctx.Event.UserID)

	rankType, num, date := analysisArgs(utils.GetArgs(ctx))
	if !cast.ToBool(proxy.GetPluginConfig("pixiv", "r18")) && strings.Contains(rankType, "r18") {
		ctx.Send("不可以涩涩！")
		return
	}
	if !utils.IsMessagePrimary(ctx) && strings.Contains(rankType, "r18") {
		ctx.Send("滚滚滚去私聊")
		return
	}
	// 获取图片信息
	pics, err := getPixivRankByHIBI(rankType, num, date)
	if err != nil {
		log.Errorf("getPixivRankByHIBI type=%v,num=%v,date=%v error: %v", rankType, num, date, err)
		ctx.Send("失败了...")
		return
	}
	if len(pics) == 0 {
		log.Warnf("No Pics : type=%v,num=%v,date=%v", rankType, num, date)
		ctx.Send("没图了...")
		return
	}
	// 开始发送
	if len(date) == 0 {
		date = "今天"
	}
	if transType, ok := translateRankTypeMap[rankType]; ok {
		rankType = transType
	}
	ctx.Send(fmt.Sprintf("开始发送%v的%v排行榜，共计%v张", date, rankType, len(pics)))
	for _, pic := range pics {
		pic.ReplaceURLToProxy() // 使用反代
		// 生成消息
		msg, err := pic.GenSinglePicMsg()
		if err != nil { // 下载图片失败
			log.Warnf("GenSinglePicMsg failed: pid=%v, err=%v", pic.PID, err)
			msg = message.Message{message.Text(pic.Title + "\n图片下载失败或无效图片\n" + pic.GetDescribe())}
		}
		ctx.Send(msg)
	}
}

var rankTypeMap = map[int]string{
	1:  "day",
	2:  "week",
	3:  "month",
	4:  "day_male",
	5:  "day_female",
	6:  "week_original",
	7:  "week_rookie",
	8:  "day_r18",
	9:  "week_r18",
	10: "day_male_r18",
	11: "day_female_r18",
}

var translateRankTypeMap = map[string]string{
	"day":            "日",
	"week":           "周",
	"month":          "月",
	"day_male":       "男性向日",
	"day_female":     "女性向日",
	"week_original":  "原创",
	"week_rookie":    "新人",
	"day_r18":        "日R-18",
	"week_r18":       "周R-18",
	"day_male_r18":   "男性向R-18",
	"day_female_r18": "女性向R-18",
}

// 分析参数
func analysisArgs(args string) (rankType string, num int, date string) {
	rankType = rankTypeMap[1]
	num = 5 // 默认5张
	args = strings.TrimSpace(args)
	if len(args) == 0 {
		return
	}
	subs := strings.Split(args, " ")
	if len(subs) >= 1 {
		tpID, err := strconv.Atoi(subs[0])
		if err != nil || tpID < 1 || tpID > len(rankTypeMap) {
			tpID = 1
		}
		rankType = rankTypeMap[tpID]
	}
	if len(subs) >= 2 {
		var err error
		num, err = strconv.Atoi(subs[1])
		if err != nil || num < 1 || num > 20 { // 最多20张
			num = 5
		}
	}
	if len(subs) >= 3 {
		_, err := time.Parse("2006-01-02", subs[2])
		if err == nil {
			date = subs[2]
		}
	}
	return
}
