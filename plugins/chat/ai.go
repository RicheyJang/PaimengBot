package chat

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func AIReply(ctx *zero.Ctx, question string) message.Message {
	// 整顿URL
	api := proxy.GetConfigString("ai.api")
	if len(api) == 0 {
		return nil
	}
	if !strings.Contains(api, "%s") {
		log.Error("AI API内必须包含%s作为问句的占位符，请重新配置")
		return nil
	}
	if !strings.HasPrefix(api, "http") {
		api = "http://" + api
	}
	api = fmt.Sprintf(api, url.QueryEscape(question))
	// 请求
	cli := client.NewHttpClient(nil)
	src, err := cli.GetReader(api)
	if err != nil {
		log.Errorf("GET %v error: %v", api, err)
		return nil
	}
	defer src.Close()
	rsp, err := io.ReadAll(src)
	if err != nil {
		log.Errorf("GET %v error: %v", api, err)
		return nil
	}
	// 解析请求
	answer := string(rsp)
	response := proxy.GetConfigString("ai.response")
	if len(response) > 0 { // json格式
		answer = gjson.Get(answer, response).String()
	}
	// 执行替换
	replaces := cast.ToStringMapString(proxy.GetConfig("ai.replaces"))
	for k, v := range replaces {
		answer = strings.ReplaceAll(answer, k, v)
	}
	answer = strings.ReplaceAll(answer, "\\n", "\n")
	// 尾处理
	if len(answer) > 0 {
		tip := proxy.GetConfigString("ai.tip")
		tipMsg := message.ParseMessageFromString(tip)
		return append(tipMsg, message.Text(answer))
	}
	return nil
}

func aiDealer(ctx *zero.Ctx, question string) message.Message {
	if !proxy.GetConfigBool("ai.enable") {
		return nil
	}
	return AIReply(ctx, question)
}
