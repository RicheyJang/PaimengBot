package push

import (
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Send 向目标推送目标消息
func Send(target Target) {
	target.Send()
}

// Target 推送目标
type Target struct {
	Msg        message.Message
	Friends    []int64
	Groups     []int64
	DoNotCheck bool

	Ctx *zero.Ctx
}

func (t *Target) GetCtx() *zero.Ctx {
	if t.Ctx == nil {
		t.Ctx = utils.GetBotCtx()
	}
	return t.Ctx
}

func (t *Target) CheckFriends() {
	list := t.GetCtx().GetFriendList().Array()
	t.Friends = intersectionWithJsonArray(t.Friends, list, "user_id")
}

func (t *Target) CheckGroups() {
	list := t.GetCtx().GetGroupList().Array()
	t.Groups = intersectionWithJsonArray(t.Groups, list, "group_id")
}

// SelfCheck 检查提供的Friends是否为自己的好友、Groups是否为自己加入的群
func (t *Target) SelfCheck() {
	t.CheckGroups()
	t.CheckFriends()
}

// Send to target.everyone
func (t *Target) Send() {
	ctx := t.GetCtx()
	if ctx == nil {
		log.Errorf("push Send: ctx is nil")
		return
	}
	if !t.DoNotCheck {
		t.SelfCheck()
	}
	if len(t.Friends)+len(t.Groups) > 0 {
		log.Infof("开始推送消息，目标私聊：%v，目标群聊：%v", t.Friends, t.Groups)
	}
	for _, friend := range t.Friends {
		ctx.SendPrivateMessage(friend, t.Msg)
	}
	for _, group := range t.Groups {
		ctx.SendGroupMessage(group, t.Msg)
	}
}
