package welcome

import (
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

var info = manager.PluginInfo{
	Name: "群欢迎消息",
	Usage: `用法：
	设置群欢迎消息 [消息...]：当有新人加群时，自动发送所设置的欢迎消息`,
	AdminLevel: 3,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"设置群欢迎消息", "自定义群欢迎消息"}, zero.OnlyGroup).
		SetBlock(true).SetPriority(3).Handle(setGroupWelcome)
	proxy.OnNotice(utils.CheckDetailType("group_increase"), func(ctx *zero.Ctx) bool {
		return ctx.Event.SelfID != ctx.Event.UserID
	}).SetBlock(false).SecondPriority().Handle(handleIncrease)
}

func setGroupWelcome(ctx *zero.Ctx) {
	var welmsg message.Message
	cmd := utils.GetCommand(ctx)
	// 消除首段消息前的Bot昵称
	first := ctx.Event.Message[0]
	first.Data["text"] = strings.TrimLeft(first.Data["text"], " ") // Trim!
	text := first.Data["text"]
	for _, nickname := range utils.GetBotConfig().NickName {
		if strings.HasPrefix(text, nickname) {
			first.Data["text"] = text[len(nickname):]
			break
		}
	}
	// 消除首段消息前的Bot昵称
	first.Data["text"] = strings.Replace(first.Data["text"], cmd, "", 1)
	first.Data["text"] = strings.TrimLeft(first.Data["text"], " ") // Trim!
	// 拼接消息
	if len(first.Data["text"]) > 0 {
		welmsg = append(welmsg, message.Text(first.Data["text"]))
	}
	for _, msg := range ctx.Event.Message[1:] {
		welmsg = append(welmsg, msg)
	}
	// 更新数据库
	preGroup := dao.GroupSetting{
		ID:      ctx.Event.GroupID,
		Welcome: welmsg.String(),
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"welcome"}), // Upsert
	}).Create(&preGroup).Error; err != nil {
		log.Errorf("set group(%v) welcome error(sql): %v", ctx.Event.GroupID, err)
		ctx.Send("设置失败了...")
		return
	}
	log.Infof("群%v的欢迎消息设置为：%v", preGroup.ID, preGroup.Welcome)
	ctx.Send("好哒，新人入群欢迎消息设置为：\n" + preGroup.Welcome)
}

// 有新人入群时
func handleIncrease(ctx *zero.Ctx) {
	var groupS dao.GroupSetting
	res := proxy.GetDB().Select("id", "welcome").Take(&groupS, ctx.Event.GroupID)
	if res.RowsAffected == 0 || len(groupS.Welcome) == 0 {
		return
	}
	ctx.SendGroupMessage(ctx.Event.GroupID, message.At(ctx.Event.UserID).String()+groupS.Welcome)
}
