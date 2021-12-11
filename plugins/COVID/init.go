package COVID

import (
	"fmt"
	"strings"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var proxy *manager.PluginProxy
var info = manager.PluginInfo{
	Name: "新冠疫情查询",
	Usage: `
查询指定地区的今日疫情状况，也可以查询全世界的
用法：
	新冠疫情 [地区]?
`,
	Classify: "实用工具",
}

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"疫情", "新冠疫情"}).SetBlock(true).ThirdPriority().Handle(covid19Handler)
}

func covid19Handler(ctx *zero.Ctx) {
	arg := strings.TrimSpace(utils.GetArgs(ctx))
	// 查询
	lines, err := GetCOVID19Condition(arg)
	if err != nil {
		log.Errorf("GetCOVID19Condition err: %v", err)
		ctx.Send("出错了...")
		return
	}
	if lines == nil || len(lines) == 0 { // 结果为空
		ctx.Send(fmt.Sprintf("%v也不知道哦", utils.GetBotNickname()))
		return
	}
	// 构造回应
	var str string
	if len(arg) > 0 {
		str = fmt.Sprintf("%s的疫情数据\n", arg)
	} else {
		str = fmt.Sprintf("全国疫情\n")
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.Text(str+strings.Join(lines, "\n")))
}

// GetCOVID19Condition 获取指定area新冠疫情状态，返回 状态项名 -> 状态值 映射
func GetCOVID19Condition(area string) ([]string, error) {
	return getCOVID19ConditionBy163(area)
}
