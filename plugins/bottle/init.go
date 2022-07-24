package bottle

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

var info = manager.PluginInfo{
	Name: "漂流瓶",
	Usage: `随风飘去或捡起他思
用法：
	扔漂流瓶 [内容]：请文明用语哦
	捡漂流瓶`,
	SuperUsage: `
	删除漂流瓶 [漂流瓶ID]：让这个漂流瓶永远消失（群管也可执行）
	删除所有漂流瓶：让当前已有的所有漂流瓶永远消失
config-plugin配置项：
	bottle.max：最多容纳多少漂流瓶，溢出时会丢弃较早放入的漂流瓶
	bottle.black：禁用词汇列表
	bottle.destroy：是(true)否(false)在捡起漂流瓶后顺便删除，默认不删`,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"扔漂流瓶", "丢漂流瓶"}).SetBlock(true).SetPriority(3).Handle(dropHandler)
	proxy.OnFullMatch([]string{"捡漂流瓶", "捡起漂流瓶"}).SetBlock(true).SetPriority(3).Handle(pickHandler)
	proxy.OnCommands([]string{"删除漂流瓶"}, zero.SuperUserPermission).SetBlock(true).SetPriority(3).Handle(deleteHandler)
	proxy.OnFullMatch([]string{"删除所有漂流瓶"}, zero.SuperUserPermission).SetBlock(true).SetPriority(3).Handle(deleteAllHandler)
	proxy.AddConfig("max", 200)
	proxy.AddConfig("black", []string{"爹", "爸"})
	proxy.AddConfig("destroy", false)
}

func dropHandler(ctx *zero.Ctx) {
	content := strings.TrimSpace(utils.GetArgs(ctx))
	if len(content) == 0 {
		ctx.Send("漂流瓶的内容怎么是空的？")
		return
	}
	if len(content) > 1000 {
		ctx.Send("内容太长啦，放不下啦")
		return
	}
	for _, word := range proxy.GetConfigStrings("black") {
		if strings.Contains(content, word) {
			ctx.Send("内容中包含禁用词汇")
			return
		}
	}
	// 记录数据库
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		// 判断是否溢出
		var count int64
		if err := tx.Model(&DriftingBottleModel{}).Count(&count).Error; err != nil {
			return err
		}
		max := proxy.GetConfigInt64("max")
		if max > 0 && count >= max { // 溢出，删除最早的
			if err := tx.Where("id IN (?)",
				tx.Model(&DriftingBottleModel{}).Select("id").Order("created_at").Limit(int(count-max)+1)).
				Delete(&DriftingBottleModel{}).Error; err != nil {
				return err
			}
		}
		// 创建
		bottle := DriftingBottleModel{
			FromID:  ctx.Event.UserID,
			Content: content,
		}
		return tx.Create(&bottle).Error
	})
	if err != nil {
		log.Errorf("sql error: %v", err)
		ctx.Send("啊这，它随风飘回来了")
		return
	}
	ctx.Send("已经替你扔出去啦")
}

func pickHandler(ctx *zero.Ctx) {
	var bottle DriftingBottleModel
	res := proxy.GetDB().Where("from_id <> ?", ctx.Event.UserID).Scopes(proxy.SQLRandomOrder).Limit(1).Find(&bottle)
	if res.Error != nil {
		log.Errorf("select error: %v", res.Error)
		ctx.Send("失败了...")
		return
	}
	if res.RowsAffected == 0 {
		ctx.Send("啥都没捞着")
		return
	}
	ctx.Send(genBottleMsg(bottle))
	if proxy.GetConfigBool("destroy") {
		proxy.GetDB().Delete(bottle)
	}
}

func deleteHandler(ctx *zero.Ctx) {
	// 校验权限
	if !(utils.IsSuperUser(ctx.Event.UserID) || auth.CheckPriority(ctx, auth.DefaultAdminLevel, false)) {
		return
	}
	// 执行
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		log.Errorf("parse id(%v) err: %v", arg, err)
		ctx.Send("ID格式错误，可以看看帮助")
		return
	}
	if err := proxy.GetDB().Where("id = ?", id).Delete(&DriftingBottleModel{}).Error; err != nil {
		log.Errorf("delete err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("好哒")
	return
}

func deleteAllHandler(ctx *zero.Ctx) {
	ctx.Send("这将删除所有的漂流瓶，是否确定？")
	event := utils.WaitNextMessage(ctx)
	if event == nil { // 无回应
		return
	}
	confirm := strings.TrimSpace(event.Message.ExtractPlainText())
	if !(confirm == "是" || confirm == "确定" || confirm == "确认") {
		ctx.Send("已取消")
		return
	}
	if err := proxy.GetDB().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&DriftingBottleModel{}).Error; err != nil {
		log.Errorf("delete err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("好哒")
	return
}

func genBottleMsg(bottle DriftingBottleModel) (msg message.MessageSegment) {
	var err error
	defer func() {
		if err != nil {
			log.Errorf("genBottleMsg err: %v", err)
			msg = message.Text(fmt.Sprintf("漂流瓶ID：%d\n%s", bottle.ID, bottle.Content))
		}
	}()
	W, H := 200.0, 50.0
	// 测量长度 分行
	img := images.NewImageCtx(1, 1)
	if err = img.UseDefaultFont(18); err != nil {
		return
	}
	var current, result string
	for _, word := range []rune(bottle.Content) {
		if word == '\n' {
			result += current + "\n"
			current = ""
			continue
		}
		if w, _ := img.MeasureString(current + string(word)); w > W {
			if current == "" {
				result += string(word) + "\n"
				current = ""
				continue
			} else {
				result += current + "\n"
				current = ""
			}
		}
		current += string(word)
	}
	if current != "" {
		result += current
	}
	newW, newH := img.MeasureMultilineString(result, 1.45)
	if newW > W {
		W = newW
	}
	if newH > H {
		H = newH
	}
	W, H = W+40, H+60
	// 画图
	img = images.NewImageCtxWithBGColor(int(W), int(H), "#faf9de")
	if err = img.UseDefaultFont(18); err != nil {
		return
	}
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawString(fmt.Sprintf("漂流瓶ID %d", bottle.ID), 10, 30)
	lines := strings.Split(result, "\n")
	y := 65.0
	for _, line := range lines {
		img.DrawString(line, 20, y-5)
		img.PasteLine(10, y, W-10, y, 2, "black")
		y += 25
	}
	msg, err = img.GenMessageAuto()
	return
}
