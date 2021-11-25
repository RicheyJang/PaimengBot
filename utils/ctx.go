package utils

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func GetArgs(ctx *zero.Ctx) string {
	res, ok := ctx.State["args"]
	if !ok {
		return ""
	}
	return cast.ToString(res)
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

// CheckDetailType 检查事件DetailType(MessageType/NoticeType/RequestType)
func CheckDetailType(tp string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if ctx.Event != nil {
			return ctx.Event.DetailType == tp
		}
		return false
	}
}
