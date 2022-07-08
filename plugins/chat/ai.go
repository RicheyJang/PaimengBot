package chat

import (
	"fmt"
	"io"
	"net/http"
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
	body := proxy.GetConfigString("ai.body")
	if !strings.Contains(api, "%s") && !strings.Contains(body, "%s") {
		log.Error("AI问答API或Body内必须包含%s作为问句的占位符，请重新配置")
		return nil
	}
	if !strings.HasPrefix(api, "http") {
		api = "http://" + api
	}
	if strings.Contains(api, "%s") {
		api = fmt.Sprintf(api, url.QueryEscape(question))
	}
	if strings.Contains(body, "%s") {
		body = fmt.Sprintf(body, question)
	}
	// 请求
	cli := client.NewHttpClient(nil)
	cli.SetUserAgent()
	var src *http.Response
	var err error
	if len(body) == 0 { // 无Body 使用GET请求
		src, err = cli.Get(api)
		if err != nil {
			log.Errorf("AI GET %v error: %v", api, err)
			return nil
		}
	} else { // 含Body 使用POST json请求
		src, err = cli.Post(api, "application/json", strings.NewReader(body))
		if err != nil {
			log.Errorf("AI POST %v error: %v", api, err)
			return nil
		}
	}
	defer src.Body.Close()
	rsp, err := io.ReadAll(src.Body)
	if err != nil {
		log.Errorf("AI read from body error: api=%v, %v", api, err)
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
		tip = strings.ReplaceAll(tip, "\\n", "\n")
		tipMsg := message.ParseMessageFromString(tip)
		answerMsg := message.ParseMessageFromString(answer)
		return append(tipMsg, answerMsg...)
	} else {
		log.Warn("AI问答API答句为空，原始回包：", string(rsp))
		return nil
	}
}

func aiDealer(ctx *zero.Ctx, question string) message.Message {
	if !proxy.GetConfigBool("ai.enable") {
		return nil
	}
	// 加锁，防止频繁调用
	if proxy.LockUser(ctx.Event.UserID) {
		return nil
	}
	defer proxy.UnlockUser(ctx.Event.UserID)
	msg := AIReply(ctx, question)
	if len(msg) > 0 {
		log.Infof("from AI API")
	}
	return msg
}
