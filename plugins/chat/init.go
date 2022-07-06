package chat

import (
	"fmt"
	"strconv"
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
如何进行自定义问答：（默认仅限群管理员在群中调用）
	新增问答
	[问句内容]
	[回答内容]：注意共三行哦，将会添加一条相应的问答，[回答内容]可以为多行
另外也可以通过单行一句话快速新增：[机器人昵称如"派蒙"]我问[问句内容]你答[回答内容]
如何查看已有自定义问答：
	已有问答
如何删除自定义问答：
	删除问答 [问句内容]：即可将相应问答删除
	删除第[编号i]个问答：删除已有问答中的第i个
另，相应问答只会在调用上述命令的群聊中生效哦
若想添加全局问答，请联系超级用户`,
	SuperUsage: `超级用户在私聊中调用上述命令，会对全局所有群和私聊生效

此外，还可通过文件批量导入问答集，可选文件格式参见DIYDialogueDir/0.txt及0.json；
文件名为生效群号，以英文逗号分隔，0代表全局生效；文件请统一放置于DIYDialogueDir目录下
txt格式文件特有：
  问句支持正则表达式（以/开头/结尾 用于标识使用正则）；
  答句支持CQ码和\n作为换行；答句支持变量：
    {bot}代表机器人昵称；{id}代表提问者ID；{nickname}代表提问者昵称
    若问句为正则，则可以通过{reg[i]}代表匹配到的问句中的第i个分组，i从1开始

config-plugin文件配置项：
	chat.default.self 自我介绍内容
	chat.default.donotknow 无法处理某消息时的回答内容，留空则不回答;可以为数组格式来随机回复;{nickname}代表机器人昵称
	chat.diylevel 自定义问答功能所需的最低管理员权限等级，默认为5，设为0则非群管理员用户也可自定义
	chat.at 在群聊中，机器人的回复是(true)否(false)@提问者
	chat.ai.enable 是(true)否(false)启用AI问答，可自行配置AI问答API
	chat.ai.api AI问答所使用的API完整网址，必须包含%s用于放置问句
	chat.ai.replaces 对AI问答的答句进行的词语替换映射
	chat.ai.response 若AI问答API的回包为json格式，则在此填写答句的字段key；为空则直接将整个回包作为答句
	chat.ai.tip 若触发AI问答，在答复消息前添加的附加前缀，可以为空`,
}

const DIYDialogueLevelKey = "diylevel"

func init() {
	info.SuperUsage = strings.ReplaceAll(info.SuperUsage, "DIYDialogueDir", consts.DIYDialogueDir)
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"新增对话", "新增问答"}).SetBlock(true).SetPriority(5).Handle(addDialogue)
	proxy.OnRegex("^我问([^\n]*)你答([^\n]*)$", zero.OnlyToMe).SetBlock(true).SetPriority(7).Handle(addDialogue)
	proxy.OnCommands([]string{"删除对话", "删除问答"}).SetBlock(true).SetPriority(5).Handle(delDialogue)
	proxy.OnRegex(`^删除第(\d+)?个问答`).SetBlock(true).SetPriority(7).Handle(delDialogue)
	proxy.OnCommands([]string{"已有对话", "已有问答"}).SetBlock(true).SetPriority(5).Handle(showDialogue)

	proxy.OnMessage(zero.OnlyToMe).SetBlock(true).SetPriority(10).Handle(dealChat)

	proxy.AddConfig("default.self", "我是派蒙，最好的伙伴！\n才不是应急食品呢")
	proxy.AddConfig("default.donotknow", "{nickname}不知道哦")
	proxy.AddConfig(DIYDialogueLevelKey, 5)
	proxy.AddConfig("at", true)

	proxy.AddConfig("ai.enable", false)
	proxy.AddConfig("ai.api", "http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s")
	proxy.AddConfig("ai.replaces", map[string]string{"小爱": "我", "{br}": "\\n"})
	proxy.AddConfig("ai.response", "content")
	proxy.AddConfig("ai.tip", "来自青云客API：\\n")
}

func addDialogue(ctx *zero.Ctx) {
	question, msg, err := analysisCtx(ctx)
	if err != nil || len(msg) == 0 || len(question) == 0 {
		ctx.Send("参数不对哦")
		log.Errorf("wrong addDialogue args :%v", err)
		return
	}
	if utils.StringRealLength(question) >= 180 {
		ctx.Send("问句太长啦，可以使用文件添加该问答")
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
	question := messageStringWithoutCmd(ctx)
	if len(question) == 0 {
		ctx.Send("参数不对哦")
		return
	}
	// 若为正则式
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) > 1 {
		index, _ := strconv.Atoi(subs[1])
		question = GetSpecQuestion(ctx.Event.GroupID, index-1)
		if len(question) == 0 {
			ctx.Send("没有这个问答")
			return
		}
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
	var subs []string
	cmd := utils.GetCommand(ctx)
	if len(cmd) > 0 { // 命令式
		// 去除消息的命令前缀
		wholeMsg := ctx.MessageString() // 完整消息（包含CQ码）
		id := strings.Index(wholeMsg, cmd)
		if id < 0 {
			return "", nil, fmt.Errorf("unexpected error: no command")
		}
		id += len(cmd)
		wholeMsg = wholeMsg[id:]
		subs = strings.SplitN(wholeMsg, "\n", 3)
	} else { // 正则式
		subs = utils.GetRegexpMatched(ctx)
	}
	// 解析命令消息
	if len(subs) < 3 {
		return "", nil, fmt.Errorf("too few parameters")
	}
	question = preprocessQuestion(strings.TrimSpace(subs[1]))
	answer = append(answer, message.ParseMessageFromString(subs[2])...)
	return
}

// 去除消息的命令前缀的CQ码格式
func messageStringWithoutCmd(ctx *zero.Ctx) string {
	cmd := utils.GetCommand(ctx)
	wholeMsg := ctx.MessageString() // 完整消息（包含CQ码）
	id := strings.Index(wholeMsg, cmd)
	if id < 0 {
		return wholeMsg
	}
	id += len(cmd)
	return strings.TrimSpace(wholeMsg[id:])
}

// 将收到的原问句org进行预处理
func preprocessQuestion(org string) string {
	if !strings.Contains(org, "[CQ:image") {
		return org
	}
	msg := message.ParseMessageFromString(org)
	for i, seg := range msg {
		if seg.Type == "image" { // 只保留file字段
			msg[i].Data = map[string]string{
				"file": seg.Data["file"],
			}
		}
	}
	return msg.String()
}
