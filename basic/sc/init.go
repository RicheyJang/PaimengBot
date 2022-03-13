package sc

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"

	"github.com/RicheyJang/PaimengBot/basic/nickname"

	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/basic/dao"

	"github.com/wdvxdr1123/ZeroBot/message"

	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
)

var info = manager.PluginInfo{
	Name: "签到与财富",
	Usage: `来签到吧！
用法；
	签到：每日签到！
	我的好感度：显示{bot}对你的好感度
	我的财富：显示你目前的资产
	好感度排行：群聊专用，显示本群{bot}好感度排行榜（前10名）
	财富排行：群聊专用，显示本群财富排行榜（前10名）`,
	SuperUsage: `
	设置好感度 [QQ号] [好感度]：摁设置指定用户的好感度
	设置财富 [QQ号] [财富]：摁设置指定用户的财富
注意，财富分为基础金额和真实金额
	基础金额为所有配置项中所体现的货币金额，用于方便各类配置
	真实金额为基础金额乘以所配置的倍率所得，用于展示给用户
config-plugin配置项：
	sc.onlygroup：是(true)否(false)只允许在群聊中签到
	sc.coin.unit: 货币单位，例如摩拉、原石
	sc.coin.rate: 基础金额和实际金额间的倍率，例如若单位为摩拉则建议设置为1000
	sc.min: 签到获得的好感度或货币(基础金额)最小值
	sc.max: 签到获得的好感度或货币(基础金额)最大值`,
}
var proxy *manager.PluginProxy

func init() {
	info.Usage = strings.ReplaceAll(info.Usage, "{bot}", utils.GetBotNickname())
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnFullMatch([]string{"签到", "每日签到"}).SetBlock(true).ThirdPriority().Handle(signHandler)
	proxy.OnFullMatch([]string{"我的好感度"}).SetBlock(true).ThirdPriority().Handle(myFavorHandler)
	proxy.OnFullMatch([]string{"我的财富", "我的资产"}).SetBlock(true).ThirdPriority().Handle(myWealthHandler)
	proxy.OnFullMatch([]string{"好感度排行", "财富排行", "好感度排行榜", "财富排行榜"}).SetBlock(true).ThirdPriority().Handle(rankHandler)
	proxy.OnCommands([]string{"设置好感度", "设置财富"}, zero.SuperUserPermission).SetBlock(true).ThirdPriority().Handle(setHandler)
	proxy.AddConfig("onlygroup", true)
	proxy.AddConfig("coin.unit", "原石")
	proxy.AddConfig("coin.rate", 80)
	proxy.AddConfig("min", 0.5)
	proxy.AddConfig("max", 1.9)
}

func myFavorHandler(ctx *zero.Ctx) {
	favor := FavorOf(ctx.Event.UserID)
	level, up := LevelAt(favor)
	ctx.Send(fmt.Sprintf("好感度：%.2f\n等级：lv%d(升级还需%.2f点好感度)", favor, level, up))
}

func myWealthHandler(ctx *zero.Ctx) {
	rc := RealCoin(BaseCoinOf(ctx.Event.UserID))
	ctx.Send(fmt.Sprintf("%v目前拥有%.0f%s",
		nickname.GetNickname(ctx.Event.UserID, "你"), rc, Unit()))
}

func signHandler(ctx *zero.Ctx) {
	// 预检查
	if proxy.GetConfigBool("onlygroup") && !utils.IsMessageGroup(ctx) {
		ctx.Send("该功能仅限在群聊中使用哦！")
		return
	}
	if proxy.LockUser(ctx.Event.UserID) {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("正在签到中~"))
		return
	}
	defer proxy.UnlockUser(ctx.Event.UserID)
	// 获取配置
	min, max := proxy.GetConfigFloat64("min"), proxy.GetConfigFloat64("max")
	if min < 0 {
		min = 0.1
	}
	if min > max {
		max = min
	}
	// 正式签到
	skipSend := false
	si := signInfo{
		id:       ctx.Event.UserID,
		name:     ctx.GetStrangerInfo(ctx.Event.UserID, false).Get("nickname").String(),
		addFavor: randomFloat(min, max),
		addCoin:  randomFloat(min, max),
	}
	if randomFloat(0, 1) > 0.95 { // 双倍
		si.double = true
		si.addCoin *= 2
		si.addFavor *= 2
	}
	// 修改数据库并判断是否已签过到
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		// 获取已有的用户数据
		user := dao.UserOwn{ID: si.id}
		res := tx.Find(&user, si.id)
		if res.Error != nil {
			return fmt.Errorf("at find | %v", res.Error)
		}
		si.orgCoin = user.Wealth
		si.orgFavor = user.Favor
		if res.RowsAffected == 0 { // 尚不存在，创建
			si.signDays = 1
			si.lastSign = time.Now()
			return tx.Create(&dao.UserOwn{
				ID:       si.id,
				Favor:    si.addFavor,
				LastSign: si.lastSign,
				SignDays: si.signDays,
				Wealth:   si.addCoin,
			}).Error
		}
		// 检查是否已经签过到
		if isSameDay(time.Now(), user.LastSign) {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("今天已经签过到了"))
			skipSend = true
			return nil
		}
		// 更新
		si.lastSign = time.Now()
		si.signDays = user.SignDays + 1
		if !isContinuouslyDay(user.LastSign, time.Now()) { // 不连续签到
			si.signDays = 1
		}
		user.SignDays = si.signDays
		user.LastSign = si.lastSign
		user.Favor += si.addFavor
		user.Wealth += si.addCoin
		return tx.Save(&user).Error
	})
	if err != nil { // 数据更新失败
		log.Errorf("[SQL] %v sign failed: %v", ctx.Event.UserID, err)
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		return
	}
	// 绘图 发送
	if !skipSend {
		ctx.Send(si.genMessage())
	}
}

func rankHandler(ctx *zero.Ctx) {
	// 判断排行榜类型
	var key string
	cmd := utils.GetCommand(ctx)
	if strings.HasPrefix(cmd, "好感") {
		key = "favor"
	} else {
		key = "wealth"
	}
	// 查询
	var users []dao.UserOwn
	if err := proxy.GetDB().Order(key + " desc").Limit(10).Find(&users).Error; err != nil {
		ctx.Send("失败了...")
		log.Errorf("[SQL] get rank error: %v", err)
		return
	}
	// 检查可行性
	if len(users) == 0 {
		ctx.Send("暂时还没有人签过到")
		return
	}
	for i, user := range users {
		if (key == "favor" && user.Favor <= 0) || (key == "wealth" && user.Wealth <= 0) {
			users = users[:i]
			break
		}
	}
	if len(users) == 0 {
		ctx.Send("暂无排名")
		return
	}
	// 绘图 发送
	msg, _ := genRankMessage(ctx, users, key)
	ctx.Send(msg)
}

func setHandler(ctx *zero.Ctx) {
	// 解析参数
	args := strings.Split(strings.TrimSpace(utils.GetArgs(ctx)), " ")
	if len(args) < 2 {
		ctx.Send("参数不足哦，可以看看帮助")
		return
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || id <= 0 {
		ctx.Send("ID格式不对哦，可以看看帮助")
		log.Warnf("ID格式错误: %v", err)
		return
	}
	value, err := strconv.ParseFloat(args[1], 64)
	if err != nil || value < 0 {
		ctx.Send("数值格式不对哦，可以看看帮助")
		log.Warnf("value格式错误: %v", err)
		return
	}
	// 修改数据库
	var orgStr string
	cmd := utils.GetCommand(ctx)
	if strings.HasSuffix(cmd, "好感度") {
		orgStr = fmt.Sprintf("%d的原好感度为%.2f，现修改为%v", id, FavorOf(id), args[1])
		err = proxy.GetDB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"favor"}),
		}).Create(&dao.UserOwn{ID: id, Favor: value}).Error
	} else { // 财富
		orgStr = fmt.Sprintf("%d的原基础货币金额为%.1f，现修改为%v(即%.1f%s)", id, BaseCoinOf(id), args[1], RealCoin(value), Unit())
		err = proxy.GetDB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"wealth"}),
		}).Create(&dao.UserOwn{ID: id, Wealth: value}).Error
	}
	if err != nil {
		log.Errorf("[SQL] update user own err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send("好哒，" + orgStr)
}

func isSameDay(a time.Time, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isContinuouslyDay(old time.Time, latest time.Time) bool {
	return isSameDay(old.AddDate(0, 0, 1), latest)
}

func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
