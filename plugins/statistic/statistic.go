package statistic

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/nickname"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "统计",
	Usage: `统计并展示各种功能使用情况
个人命令：（统计你自己的使用情况）
	统计
	今日统计
群命令：（统计本群的使用情况）
	群统计
	群今日统计
`,
	SuperUsage: `
	全局统计
	全局今日统计
统计某人的使用情况：
	统计 [QQ号]
	今日统计 [QQ号]
配置项：
	statistic.ignore: 不纳入统计范围的插件Key列表`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	manager.AddPostHook(statisticHook)
	proxy.OnCommands([]string{"统计", "今日统计"}, zero.OnlyToMe).SetBlock(true).
		SetPriority(4).Handle(selfStatistics)
	proxy.OnCommands([]string{"群统计", "群今日统计"}, zero.OnlyToMe, zero.OnlyGroup).
		SetBlock(true).SetPriority(4).Handle(groupStatistics)
	proxy.OnCommands([]string{"全局统计", "全局今日统计"}, zero.SuperUserPermission, zero.OnlyPrivate).
		SetBlock(true).SetPriority(4).Handle(globalStatistics)
	_, _ = proxy.AddScheduleDailyFunc(0, 1, initialDailyStatistics)
	proxy.AddConfig("ignore", []string{"statistic"})
	initialDailyStatistics()
}

const statisticPrefix = "statis"

// key为四段式：statisticPrefix.类型(总计g 或 当天d).ID(g群号 或 p用户ID 或 全局a).插件Key
// 值为调用次数，uint32类型

func statisticHook(condition *manager.PluginCondition, ctx *zero.Ctx) error {
	if !utils.IsMessage(ctx) { // 只记录消息型调用
		return nil
	}
	if !utils.GetNeedStatistic(ctx) { // 不需要统计
		return nil
	}
	key := condition.Key
	batch := new(leveldb.Batch)
	// 记录群
	if ctx.Event.GroupID != 0 && utils.IsMessageGroup(ctx) { // 为群消息
		dailyKey := fmt.Sprintf("%v.d.g%v.%v", statisticPrefix, ctx.Event.GroupID, key)
		sumKey := fmt.Sprintf("%v.g.g%v.%v", statisticPrefix, ctx.Event.GroupID, key)
		putKVNum(batch, dailyKey, getKVNum(dailyKey)+1)
		putKVNum(batch, sumKey, getKVNum(sumKey)+1)
	}
	// 记录个人
	if ctx.Event.UserID != 0 && !utils.IsGroupAnonymous(ctx) { // 不是群匿名消息
		dailyKey := fmt.Sprintf("%v.d.p%v.%v", statisticPrefix, ctx.Event.UserID, key)
		sumKey := fmt.Sprintf("%v.g.p%v.%v", statisticPrefix, ctx.Event.UserID, key)
		putKVNum(batch, dailyKey, getKVNum(dailyKey)+1)
		putKVNum(batch, sumKey, getKVNum(sumKey)+1)
	}
	// 记录全局
	dailyKey := fmt.Sprintf("%v.d.a.%v", statisticPrefix, key)
	sumKey := fmt.Sprintf("%v.g.a.%v", statisticPrefix, key)
	putKVNum(batch, dailyKey, getKVNum(dailyKey)+1)
	putKVNum(batch, sumKey, getKVNum(sumKey)+1)
	// 写入
	err := proxy.GetLevelDB().Write(batch, nil)
	if err != nil {
		log.Warnf("<%v> 统计记录失败，err=%v", key, err)
	}
	return nil
}

func initialDailyStatistics() {
	// 校验日期
	dateKey := fmt.Sprintf("%v.last.day", statisticPrefix)
	ls := getKVNum(dateKey)
	now := uint32(time.Now().YearDay())
	if ls == now {
		return
	}
	// 遍历所有前日调用记录
	count := 0
	batch := new(leveldb.Batch)
	putKVNum(batch, dateKey, now)
	prefix := fmt.Sprintf("%v.d.", statisticPrefix)
	iter := proxy.GetLevelDB().NewIterator(levelutil.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		batch.Delete(iter.Key())
		count += 1
	}
	iter.Release()
	// 更新
	err := proxy.GetLevelDB().Write(batch, nil)
	if err != nil {
		log.Warnf("统计记录每日初始化失败，err=%v", err)
	} else {
		log.Infof("统计记录每日初始化成功，涉及K-V：%v条", count)
	}
}

func isDaily(ctx *zero.Ctx) bool {
	cmd := utils.GetCommand(ctx)
	return strings.Contains(cmd, "今日")
}

func selfStatistics(ctx *zero.Ctx) {
	if utils.IsGroupAnonymous(ctx) {
		ctx.Send("请关闭匿名哦")
		return
	}
	userID := ctx.Event.UserID
	nick := nickname.GetNickname(userID, strconv.FormatInt(userID, 10))
	if utils.IsSuperUser(ctx.Event.UserID) { // 超级用户统计指定用户
		args := strings.TrimSpace(utils.GetArgs(ctx))
		id, err := strconv.ParseInt(args, 10, 64)
		if len(args) > 0 && err == nil {
			userID = id
			nick = args
		}
	}
	title := fmt.Sprintf("%v的功能使用量统计", nick)
	prefix := fmt.Sprintf("%v.g.p%v.", statisticPrefix, userID)
	if isDaily(ctx) {
		title = fmt.Sprintf("%v的今日功能使用量统计", nick)
		prefix = fmt.Sprintf("%v.d.p%v.", statisticPrefix, userID)
	}
	ctx.Send(dealStatistic(title, prefix))
}

func groupStatistics(ctx *zero.Ctx) {
	title := "群功能使用量统计"
	prefix := fmt.Sprintf("%v.g.g%v.", statisticPrefix, ctx.Event.GroupID)
	if isDaily(ctx) {
		title = "群功能今日使用量统计"
		prefix = fmt.Sprintf("%v.d.g%v.", statisticPrefix, ctx.Event.GroupID)
	}
	ctx.Send(dealStatistic(title, prefix))
}

func globalStatistics(ctx *zero.Ctx) {
	title := "全局功能使用量统计"
	prefix := fmt.Sprintf("%v.g.a.", statisticPrefix)
	if isDaily(ctx) {
		title = "全局功能今日使用量统计"
		prefix = fmt.Sprintf("%v.d.a.", statisticPrefix)
	}
	ctx.Send(dealStatistic(title, prefix))
}

func dealStatistic(title string, prefix string) message.MessageSegment {
	resMap := make(map[string]uint32)
	iter := proxy.GetLevelDB().NewIterator(levelutil.BytesPrefix([]byte(prefix)), nil)
	skips := proxy.GetConfigStrings("ignore")
	for iter.Next() {
		// 过滤
		num := BytesToUInt32(iter.Value())
		if num == 0 || len(iter.Key()) <= len(prefix) {
			continue
		}
		key := string(iter.Key()[len(prefix):])
		plugin := manager.GetPluginConditionByKey(key)
		if plugin == nil {
			continue
		}
		if utils.StringSliceContain(skips, plugin.Key) {
			continue
		}
		// 记录
		resMap[plugin.Name] = num
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Warnf("iter err, prefix=%v, err=%v", prefix, err)
		return message.Text("统计失败了，稍后再试试")
	}
	if len(resMap) == 0 {
		return message.Text(fmt.Sprintf("%s：\n暂时没有使用记录哦，多陪陪%v嘛", title, utils.GetBotNickname()))
	}
	// 绘图
	return drawGraph(title, resMap)
}

func getKVNum(key string) uint32 {
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		return 0
	}
	return BytesToUInt32(v)
}

func putKVNum(batch *leveldb.Batch, key string, value uint32) {
	batch.Put([]byte(key), UInt32ToBytes(value))
}

func UInt32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func BytesToUInt32(b []byte) uint32 {
	for len(b) < 4 {
		b = append(b, byte(0))
	}
	return binary.LittleEndian.Uint32(b)
}

func drawGraph(title string, mp map[string]uint32) message.MessageSegment {
	// 初始化
	var sum float64
	var values []chart.Value
	text := title + "\n"
	for k, v := range mp {
		values = append(values, chart.Value{
			Label: fmt.Sprintf("%v(%v次)", k, v),
			Value: float64(v),
		})
		text += fmt.Sprintf("%v: %v次\n", k, v)
		sum += float64(v)
	}
	text = strings.TrimSpace(text)
	// 排序
	sort.Slice(values, func(i, j int) bool {
		if values[i].Value == values[j].Value {
			return values[i].Label < values[j].Label
		}
		return values[i].Value > values[j].Value
	})
	if len(values) >= 10 { // 筛选总量小于10%的
		var part float64
		for i := len(values) - 1; i >= 0; i-- {
			part += values[i].Value
			if part/sum >= 0.1 { // 占比总量超过10%
				if i < len(values)-2 { // 至少已包含两项
					values = values[:i+1]
					values = append(values, chart.Value{
						Label: fmt.Sprintf("其它(共%v次)", int(part)),
						Value: part,
					})
				}
				break
			}
		}
	}
	// 画图
	Size := 550
	if len(values) > 6 { // 依据条数，动态调整大小
		Size += len(values) * 35
	}
	img := images.NewImageCtxWithBGColor(Size, Size, "white")
	if err := img.FillDonutChartDefault(title, values); err != nil {
		log.Warnf("FillDonutChartDefault err: %v", err)
		return message.Text(text)
	}
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("GenMessageAuto err: %v", err)
		return message.Text(text)
	}
	return msg
}
