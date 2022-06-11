package github

import (
	"bytes"
	"encoding/base64"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "GitHub查询",
	Usage: `
查询某个仓库
用法：
	github[查询名称]
例子：
	github RicheyJang/PaimengBot
`,
	Classify: "一般功能",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex("^github\\s*(.+)").SetBlock(true).SetPriority(4).Handle(handleReg)
	proxy.AddConfig("maxResult", 2)
}

const githubAPI = "https://api.github.com/search/repositories"

func handleReg(ctx *zero.Ctx) {
	// 最大结果数
	var maxResult = int(proxy.GetConfigInt64("maxResult"))
	var c = client.NewHttpClient(nil)
	target := utils.GetRegexpMatched(ctx)[1]
	// 获取结果
	json, errApi := c.GetGJson(githubAPI + "?q=" + target)
	if errApi != nil {
		log.Errorf("GitHubApi返回解析错误,err=%s", errApi.Error())
		return
	}
	if count := json.Get("total_count").Int(); count <= 0 {
		ctx.Send("没有找到这样的仓库~")
		return
	}

	var names []string
	for i, result := range json.Get("items").Array() {
		if i > maxResult {
			break
		}
		names = append(names, result.Get("full_name").String())
	}
	sendImageChain(ctx, names)
}

func sendImageChain(ctx *zero.Ctx, fullNames []string) {
	var messages []message.MessageSegment
	for _, name := range fullNames {
		messages = append(messages, getImage(name))                           // 图片
		messages = append(messages, message.Text("https://github.com/"+name)) // 仓库地址
	}
	ctx.SendChain(messages...)
}

// 获取image消息
func getImage(fullName string) message.MessageSegment {
	var c = client.NewHttpClient(nil)
	var respImage, _ = c.Get("https://opengraph.githubassets.com/0/" + fullName)
	// Base64
	resultBuff := bytes.NewBuffer(nil) // 结果缓冲区
	// 新建Base64编码器（Base64结果写入结果缓冲区resultBuff）
	encoder := base64.NewEncoder(base64.StdEncoding, resultBuff)
	// 将文件写入Base64编码器
	_, err := io.Copy(encoder, respImage.Body)
	if err != nil {
		log.Warnf("localImage2Send base64 copy err: %v", err)
		_ = encoder.Close()
		return message.Text("[Error]生成图片失败")
	}
	// 结束Base64编码
	err = encoder.Close()
	if err != nil {
		log.Warnf("localImage2Send base64 encode err: %v", err)
		return message.Text("[Error]生成图片失败")
	}
	return message.Image("base64://" + resultBuff.String())
}
