package chat

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "聊天",
	Usage: `来闲聊吧！
	如何编辑自定义对话：（仅限群管理员在群中调用）
	新增对话
	[问句内容]
	[回答内容]：注意共三行哦，将会添加一条相应的对话
	如何删除自定义对话：
	删除对话 [问句内容]：即可将相应对话删除
	另，相应对话只会在调用上述命令的群聊中生效哦
	若想添加全局对话，请联系超级用户`,
	SuperUsage: `超级用户在私聊中调用上述命令，会对全局所有群和私聊生效`,
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"新增对话", "新增问答"}, zero.OnlyToMe).SetBlock(true).SetPriority(5).Handle(addDialogue)
	proxy.OnCommands([]string{"删除对话", "删除问答"}, zero.OnlyToMe).SetBlock(true).SetPriority(5).Handle(delDialogue)
	proxy.OnCommands([]string{"已有对话", "已有问答"}, zero.OnlyToMe).SetBlock(true).SetPriority(5).Handle(showDialogue)
	proxy.OnMessage(zero.OnlyToMe).SetBlock(true).SetPriority(10).Handle(dealChat)
	proxy.AddConfig("default.self", "我是派蒙，最好的伙伴！\n才不是应急食品呢")
}

func addDialogue(ctx *zero.Ctx) {
	question, msg, err := analysisCtx(ctx)
	if err != nil || len(msg) == 0 || len(question) == 0 {
		ctx.Send("参数不对哦")
		log.Errorf("wrong addDialogue args :%v", err)
		return
	}
	if utils.IsSuperUser(ctx.Event.UserID) && utils.IsMessagePrimary(ctx) {
		// 超级用户 in 私聊
		err = SetDialogue(0, question, msg)
		if err != nil {
			log.Errorf("SetDialogue failed: question=%v, err=%v", question, err)
			ctx.Send("失败了...")
		} else {
			ctx.SendChain(append(message.Message{
				message.Text(fmt.Sprintf("新增问答：\n问：%v\n答：", question))}, msg...)...)
		}
	} else if utils.IsMessageGroup(ctx) {
		if !auth.CheckPriority(ctx, 5, true) { // 无权限
			return
		}
		// 管理员 in 群聊
		err = SetDialogue(ctx.Event.GroupID, question, msg)
		if err != nil {
			log.Errorf("SetDialogue failed: group=%v, question=%v, err=%v", ctx.Event.GroupID, question, err)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		} else {
			ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID),
				message.Text(fmt.Sprintf("新增问答：\n问：%v\n答：", question))}, msg...)...)
		}
	} else { // 其它
		ctx.Send("请在群聊中使用本功能哦，可以看看帮助")
	}
}

func delDialogue(ctx *zero.Ctx) {
	question := strings.TrimSpace(utils.GetArgs(ctx))
	if len(question) == 0 {
		ctx.Send("参数不对哦")
		return
	}
	if utils.IsSuperUser(ctx.Event.UserID) && utils.IsMessagePrimary(ctx) {
		// 超级用户 in 私聊
		err := DeleteDialogue(0, question)
		if err != nil {
			log.Errorf("DeleteDialogue failed: question=%v, err=%v", question, err)
			ctx.Send("失败了...")
		} else {
			ctx.Send("好哒")
		}
	} else if utils.IsMessageGroup(ctx) {
		if !auth.CheckPriority(ctx, 5, true) { // 无权限
			return
		}
		// 管理员 in 群聊
		err := DeleteDialogue(ctx.Event.GroupID, question)
		if err != nil {
			log.Errorf("DeleteDialogue failed: group=%v, question=%v, err=%v", ctx.Event.GroupID, question, err)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("失败了..."))
		} else {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("好哒"))
		}
	} else { // 其它
		ctx.Send("请在群聊中使用本功能哦，可以看看帮助")
	}
}

func showDialogue(ctx *zero.Ctx) {
	qs := GetAllQuestion(ctx.Event.GroupID)
	if len(qs) == 0 {
		ctx.Send("暂无自定义问答")
		return
	}
	if (utils.IsSuperUser(ctx.Event.UserID) && utils.IsMessagePrimary(ctx)) ||
		(utils.IsMessageGroup(ctx) && auth.CheckPriority(ctx, 5, true)) {
		str := "已有问题："
		for i, q := range qs {
			str += fmt.Sprintf("\n[%d] %v", i+1, q)
		}
		ctx.Send(str)
	} else {
		ctx.Send("请在群聊中使用本功能哦，可以看看帮助")
	}
}

func analysisCtx(ctx *zero.Ctx) (question string, answer message.Message, err error) {
	// 去除消息[0]的命令前缀
	cmd := utils.GetCommand(ctx)
	firstMsg := ctx.Event.Message[0].Data["text"]
	id := strings.Index(firstMsg, cmd)
	if id < 0 {
		return "", nil, fmt.Errorf("unexpected error: no command")
	}
	id += len(cmd)
	firstMsg = firstMsg[id:]
	// 解析命令消息
	subs := strings.SplitN(firstMsg, "\n", 3)
	if len(subs) < 3 {
		return "", nil, fmt.Errorf("too few parameters")
	}
	question = strings.TrimSpace(subs[1])
	answer = append(answer, message.Text(subs[2]))
	if len(ctx.Event.Message) > 1 {
		answer = append(answer, ctx.Event.Message[1:]...)
	}
	return
}
