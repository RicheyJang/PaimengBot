package utils

import (
	"fmt"
	"io"
	"strconv"
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

// WaitNextMessage 等待相同用户的下一条消息，若返回nil，代表超时（5分钟）
func WaitNextMessage(ctx *zero.Ctx) *zero.Event {
	t := time.NewTimer(5 * time.Minute)
	defer t.Stop()
	r, cancel := ctx.FutureEvent("message", ctx.CheckSession()).Repeat()
	defer cancel()
	select {
	case e := <-r:
		return e
	case <-t.C: // 超时取消
		return nil
	}
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

// GetQQAvatar 快捷获取QQ头像
func GetQQAvatar(qq int64, size int) (io.Reader, error) {
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
func GetQQGroupAvatar(id int64, size int) (io.Reader, error) {
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

// IsSuperUser userID是否为超级用户
func IsSuperUser(userID int64) bool {
	uid := strconv.FormatInt(userID, 10)
	for _, su := range GetBotConfig().SuperUsers {
		if su == uid {
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
			userID, err := strconv.ParseInt(user, 10, 64)
			if err != nil {
				continue
			}
			ctx.SendPrivateMessage(userID, message)
		}
		return true
	})
}

// ---- Rules ----

// CheckDetailType Rule:检查事件DetailType(MessageType/NoticeType/RequestType)
func CheckDetailType(tp string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if ctx.Event != nil {
			return ctx.Event.DetailType == tp
		}
		return false
	}
}

// SkipGroupAnonymous Rule:不处理群匿名消息
func SkipGroupAnonymous(ctx *zero.Ctx) bool {
	return !IsGroupAnonymous(ctx)
}

// SkipGuildMessage Rule:不处理频道消息事件
func SkipGuildMessage(ctx *zero.Ctx) bool {
	return !IsMessageGuild(ctx)
}
