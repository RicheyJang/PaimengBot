package chat

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/auth"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "聊天",
	Usage: `来闲聊吧！
	如何进行自定义问答：（默认仅限群管理员在群中调用，且开头需要加上机器人昵称 如派蒙）
	新增问答
	[问句内容]
	[回答内容]：注意共三行哦，将会添加一条相应的问答，[回答内容]可以为多行
	另外也可以通过单行一句话快速新增：[机器人昵称如"派蒙"]我问[问句内容]你答[回答内容]
	如何删除自定义问答：
	删除问答 [问句内容]：即可将相应问答删除
	如何查看已有自定义问答：
	已有问答
	另，相应问答只会在调用上述命令的群聊中生效哦
	若想添加全局问答，请联系超级用户`,
	SuperUsage: `超级用户在私聊中调用上述命令，会对全局所有群和私聊生效
此外，还可通过文件批量导入问答集，可选文件格式参见DIYDialogueDir/0.txt及0.json
文件名为生效群号，以英文逗号分隔，0代表全局生效；文件请统一放置于DIYDialogueDir目录下
config-plugin文件配置项：
chat.default.self 自我介绍内容
chat.diylevel 自定义问答功能所需的最低管理员权限等级，默认为5，设为0则非群管理员用户也可自定义
chat.onlytome 在群中调用已自定义的问句时是(true)否(false)需要加上机器人名字前缀或者@机器人
chat.at 在群聊中，机器人的回复是(true)否(false)@提问者`,
}

const DIYDialogueLevelKey = "diylevel"

func init() {
	info.SuperUsage = strings.ReplaceAll(info.SuperUsage, "DIYDialogueDir", consts.DIYDialogueDir)
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"新增对话", "新增问答"}).SetBlock(true).SetPriority(5).Handle(addDialogue)
	proxy.OnRegex("我问([^\n]*)你答([^\n]*)", zero.OnlyToMe).SetBlock(true).SetPriority(7).Handle(addDialogue)
	proxy.OnCommands([]string{"删除对话", "删除问答"}).SetBlock(true).SetPriority(5).Handle(delDialogue)
	proxy.OnCommands([]string{"已有对话", "已有问答"}).SetBlock(true).SetPriority(5).Handle(showDialogue)

	proxy.OnMessage().SetBlock(true).SetPriority(10).Handle(dealChat)

	proxy.AddConfig("default.self", "我是派蒙，最好的伙伴！\n才不是应急食品呢")
	proxy.AddConfig(DIYDialogueLevelKey, 5)
	proxy.AddConfig("onlytome", true)
	proxy.AddConfig("at", true)
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
		if !auth.CheckPriority(ctx, int(proxy.GetConfigInt64(DIYDialogueLevelKey)), true) { // 无权限
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
		if !auth.CheckPriority(ctx, int(proxy.GetConfigInt64(DIYDialogueLevelKey)), true) { // 无权限
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
		(utils.IsMessageGroup(ctx) &&
			auth.CheckPriority(ctx, int(proxy.GetConfigInt64(DIYDialogueLevelKey)), true)) {
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
		subs = utils.GetRegexpMatched(ctx)
	}
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
