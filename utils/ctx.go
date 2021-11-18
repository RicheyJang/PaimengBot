package utils

import (
	"strconv"

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

// GetBotConfig 获取机器人配置
func GetBotConfig() zero.Config {
	return zero.BotConfig
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
