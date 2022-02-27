package nickname

import (
	"fmt"
	"regexp"
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
	nickname.blackname: 黑名单词列表，禁止用户使用这些词作为昵称`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex("(以后)?叫我.+", zero.OnlyToMe, rules.SkipGroupAnonymous).
		SetBlock(true).SecondPriority().Handle(setNickName)
	proxy.AddConfig("blackName", []string{"爸", "妈", "爹", "主人"})
}

func setNickName(ctx *zero.Ctx) {
	reg := regexp.MustCompile("(以后)?叫我(.+)")
	sub := reg.FindStringSubmatch(ctx.MessageString())
	if len(sub) < 3 {
		ctx.Send("喂！你倒是告诉我叫你什么呀")
		return
	}
	nick := sub[2]
	if !utils.IsSuperUser(ctx.Event.UserID) { // 非超级用户，需要判断昵称黑名单
		blackName := append(proxy.GetConfigStrings("blackName"), utils.GetBotConfig().NickName...)
		for _, black := range blackName {
			if strings.Contains(nick, black) {
				ctx.Send("才不叫你这个呢！")
				return
			}
		}
	}
	userS := dao.UserSetting{ID: ctx.Event.UserID, Nickname: nick}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname"}), // Upsert
	}).Create(&userS).Error; err != nil {
		log.Errorf("set nickname error(sql): %v", err)
		ctx.Send("设置昵称失败了哦")
	} else {
		ctx.Send(fmt.Sprintf("好哒，以后就叫你%v咯", nick))
	}
}
