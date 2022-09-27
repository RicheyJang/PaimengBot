package utils

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// GetArgs 获取参数
func GetArgs(ctx *zero.Ctx) string {
	if ctx == nil {
		return ""
	}
	res, ok := ctx.State["args"]
	if !ok {
		return ""
	}
	return cast.ToString(res)
}

// GetCommand 获取命令
func GetCommand(ctx *zero.Ctx) string {
	if ctx == nil {
		return ""
	}
	res, ok := ctx.State["command"]
	if !ok { // 尝试全匹配
		res, ok = ctx.State["matched"]
	}
	if !ok {
		return ""
	}
	return cast.ToString(res)
}

// GetRegexpMatched 获取正则匹配字符串切片
func GetRegexpMatched(ctx *zero.Ctx) []string {
	if ctx == nil {
		return nil
	}
	res, ok := ctx.State["regex_matched"]
	if !ok {
		return nil
	}
	return cast.ToStringSlice(res)
}

// SetNotStatistic 设置此次调用不统计
func SetNotStatistic(ctx *zero.Ctx) {
	if ctx == nil {
		return
	}
	ctx.State["do_not_statistic"] = true
}

// GetNeedStatistic 获取此次调用是否需要统计
func GetNeedStatistic(ctx *zero.Ctx) bool {
	if ctx == nil { // 默认统计
		return true
	}
	res, ok := ctx.State["do_not_statistic"]
	if !ok {
		return true
	}
	return !cast.ToBool(res)
}

// WaitNextMessage 等待相同用户的下一条消息，若返回nil，代表超时（5分钟）
func WaitNextMessage(ctx *zero.Ctx) *zero.Event {
	t := time.NewTimer(5 * time.Minute)
	defer t.Stop()
	r, cancel := ctx.FutureEvent("message", ctx.CheckSession()).Repeat()
	defer cancel()
	select {
	case e := <-r:
		return e.Event
	case <-t.C: // 超时取消
		return nil
	}
}

// GetConfirm 发送tip并获取用户确认消息，为true代表用户确定
func GetConfirm(tip string, ctx *zero.Ctx) bool {
	ctx.Send(tip)
	event := WaitNextMessage(ctx)
	if event == nil { // 无回应
		return false
	}
	confirm := strings.TrimSpace(event.Message.ExtractPlainText())
	if confirm == "是" || confirm == "确定" || confirm == "确认" {
		return true
	}
	return false
}

// GetImageURL 通过消息获取其中的图片URL
func GetImageURL(msg message.MessageSegment) string {
	if msg.Type != "image" {
		return ""
	}
	return msg.Data["url"]
}

// GetImageURLs 获取消息全部图片URL
func GetImageURLs(e *zero.Event) (urls []string) {
	if e == nil {
		return
	}
	for _, msg := range e.Message {
		url := GetImageURL(msg)
		if len(url) > 0 {
			urls = append(urls, url)
		}
	}
	return
}

// IsMessage 是否为消息事件
func IsMessage(ctx *zero.Ctx) bool {
	if ctx == nil || ctx.Event == nil {
		return false
	}
	return ctx.Event.PostType == "message"
}

// IsMessagePrimary 是否为私聊消息
func IsMessagePrimary(ctx *zero.Ctx) bool {
	if ctx == nil || ctx.Event == nil {
		return false
	}
	return ctx.Event.PostType == "message" && ctx.Event.MessageType == "private"
}

// IsMessageGroup 是否为群聊消息
func IsMessageGroup(ctx *zero.Ctx) bool {
	if ctx == nil || ctx.Event == nil {
		return false
	}
	return ctx.Event.PostType == "message" && ctx.Event.MessageType == "group"
}

// IsMessageGuild 是否为频道消息
func IsMessageGuild(ctx *zero.Ctx) bool {
	if ctx == nil || ctx.Event == nil {
		return false
	}
	return ctx.Event.PostType == "message" && ctx.Event.MessageType == "guild"
}

// IsGroupAnonymous 判断是否为群匿名消息
func IsGroupAnonymous(ctx *zero.Ctx) bool {
	if !IsMessageGroup(ctx) {
		return false
	}
	return ctx.Event.SubType == "anonymous"
}

// IsGroupAdmin 是否为群管消息
func IsGroupAdmin(ctx *zero.Ctx) bool {
	if !IsMessageGroup(ctx) || ctx.Event.Sender == nil {
		return false
	}
	return ctx.Event.Sender.Role == "owner" || ctx.Event.Sender.Role == "admin"
}

// GetQQAvatar 快捷获取QQ头像
func GetQQAvatar(qq int64, size int) (io.ReadCloser, error) {
	c := client.NewHttpClient(&client.HttpOptions{
		Timeout: 3 * time.Second,
		TryTime: 2,
	})
	url1 := fmt.Sprintf("http://q1.qlogo.cn/g?b=qq&nk=%v&s=%v", qq, size)
	url2 := fmt.Sprintf("https://q2.qlogo.cn/headimg_dl?dst_uin=%v&spec=%v", qq, size)
	res, err := c.GetReader(url2) // 尝试q2
	if err != nil {
		res, err = c.GetReader(url1) // 失败则尝试q1
		if err != nil {
			log.Errorf("获取QQ头像失败, err: %v", err)
			return nil, err
		}
	}
	return res, err
}

// GetQQGroupAvatar 快捷获取QQ群头像
func GetQQGroupAvatar(id int64, size int) (io.ReadCloser, error) {
	c := client.NewHttpClient(&client.HttpOptions{
		Timeout: 3 * time.Second,
		TryTime: 2,
	})
	url := fmt.Sprintf("http://p.qlogo.cn/gh/%v/%v/%v/", id, id, size)
	res, err := c.GetReader(url) // 尝试
	if err != nil {
		log.Errorf("获取QQ群头像失败, err: %v", err)
		return nil, err
	}
	return res, err
}

var tmpAddressBuff = make(map[string]bool)

// IsOneBotLocal 判断OneBot(消息收发端)是否在本地
func IsOneBotLocal() (res bool) {
	addr := viper.GetString("server.address")
	defer func() {
		tmpAddressBuff[addr] = res
	}()
	if res, ok := tmpAddressBuff[addr]; ok { // 读取缓存
		return res
	}
	sub := strings.Split(addr, "//")
	if len(sub) < 2 {
		return false
	}
	if strings.HasPrefix(sub[1], "127") || strings.HasPrefix(sub[1], "local") {
		return true
	}
	return false
}

// GetBotCtx 获取一个全局ctx
func GetBotCtx() *zero.Ctx {
	var res *zero.Ctx
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		if ctx != nil {
			res = ctx
			return false
		}
		return true
	})
	return res
}

// GetBotConfig 获取机器人配置
func GetBotConfig() zero.Config {
	return zero.BotConfig
}

// GetBotNickname 获取机器人昵称
func GetBotNickname() string {
	nick := GetBotConfig().NickName
	if len(nick) == 0 || len(nick[0]) == 0 {
		return "我"
	}
	return nick[0]
}

// GetBotID 获取机器人的登录ID
func GetBotID() int64 {
	if len(zero.BotConfig.Driver) == 0 {
		return 0
	}
	return zero.BotConfig.Driver[0].SelfID()
}

// IsSuperUser userID是否为超级用户
func IsSuperUser(userID int64) bool {
	for _, su := range GetBotConfig().SuperUsers {
		if su == userID {
			return true
		}
	}
	return false
}

// SendToSuper 将消息发送给所有后端的所有超级用户
func SendToSuper(message ...message.MessageSegment) {
	supers := GetBotConfig().SuperUsers
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		for _, user := range supers {
			ctx.SendPrivateMessage(user, message)
		}
		return true
	})
}
