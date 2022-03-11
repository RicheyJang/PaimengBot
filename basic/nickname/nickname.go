package nickname

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/rules"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm/clause"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "昵称",
	Usage: `
用法：
	以后叫我XXX：将你的昵称设置为XXX
`,
	SuperUsage: `config-plugin配置项：
	nickname.blackname: 黑名单词列表，禁止用户昵称包含这些词
	nickname.max: 昵称的最大长度`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex("^(以后)?叫我(.+)", zero.OnlyToMe, rules.SkipGroupAnonymous).
		SetBlock(true).SecondPriority().Handle(setNickName)
	proxy.AddConfig("blackName", []string{"爸", "妈", "爹", "主人"})
	proxy.AddConfig("max", 10)
}

func setNickName(ctx *zero.Ctx) {
	sub := utils.GetRegexpMatched(ctx)
	if len(sub) <= 2 {
		ctx.Send("喂！你倒是告诉我叫你什么呀")
		return
	}
	nick := sub[2]
	// 检查长度
	if utils.StringRealLength(nick) > int(proxy.GetConfigInt64("max")) {
		ctx.Send(fmt.Sprintf("喂！你的昵称太长了，最多%v个字", proxy.GetConfigInt64("max")))
		return
	}
	// 检查黑名单
	if !utils.IsSuperUser(ctx.Event.UserID) { // 非超级用户，需要判断昵称黑名单
		blackName := append(proxy.GetConfigStrings("blackName"), utils.GetBotConfig().NickName...)
		for _, black := range blackName {
			if strings.Contains(nick, black) {
				ctx.Send("才不叫你这个呢！")
				return
			}
		}
	}
	// 设置
	userS := dao.UserSetting{ID: ctx.Event.UserID, Nickname: nick}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname"}), // Upsert
	}).Create(&userS).Error; err != nil {
		log.Errorf("set nickname error(sql): %v", err)
		ctx.Send("失败了...")
	} else {
		ctx.Send(fmt.Sprintf("好哒，以后就叫你%v咯", nick))
	}
}
