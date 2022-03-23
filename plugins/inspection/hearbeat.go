package inspection

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/push"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var heartbeatOnce sync.Once

func heartbeatConfigHook(_ fsnotify.Event) error {
	receiver := proxy.GetConfigStrings("heartbeat.receiver")
	if len(receiver) > 0 {
		heartbeatOnce.Do(func() {
			go heartbeatAndSend()
			log.Infof("开启心跳检测并向%v发送", receiver)
		})
	}
	return nil
}

func heartbeatAndSend() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("heartbeatAndSend err: %v\n%v", err, string(debug.Stack()))
		}
	}()
	for {
		if utils.GetBotCtx() == nil {
			time.Sleep(10 * time.Second)
		}
		// 检查时间段
		var from, to int
		_, _ = fmt.Sscanf(proxy.GetConfigString("heartbeat.period"), "%d-%d", &from, &to)
		if to < from {
			to = from
		}
		now := time.Now().Hour()
		if now < from || now > to {
			waitNextHeartbeat()
			continue
		}
		// 检测
		self := "机器人状态正常"
		bot := CheckOnebot(true)
		// 生成消息
		msg := message.Text("心跳消息：" + self + "，" + bot + "。\n查看详情请说\"自检\"")
		// 发送
		target := push.Target{Msg: message.Message{msg}}
		fs := proxy.GetConfigStrings("heartbeat.receiver")
		for _, f := range fs {
			target.Friends = append(target.Friends, cast.ToInt64(f))
		}
		target.Send()
		// 等待
		waitNextHeartbeat()
	}
}

func waitNextHeartbeat() {
	intervalStr := proxy.GetConfigString("heartbeat.interval")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil || interval <= time.Minute {
		log.Warnf("心跳时间间隔inspection.heartbeat.interval格式错误或过短，重置为1分钟，err=%v", err)
		interval = time.Minute
	}
	time.Sleep(interval)
}
