package sc

import (
	"fmt"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type signInfo struct {
	id       int64
	orgFavor float64
	addFavor float64
	orgCoin  float64
	addCoin  float64
	signDays int
	lastSign time.Time // 最近一次签到时间
}

func (s signInfo) genMessage() message.Message {
	// TODO 绘图
	str := fmt.Sprintf("签到成功\n已连续签到%v天\n好感度：%.2f(+%.2f)\n财富：%.0f(+%.0f)%s",
		s.signDays,
		s.orgFavor+s.addFavor, s.addFavor,
		RealCoin(s.orgCoin+s.addCoin), RealCoin(s.addCoin), Unit())
	return message.Message{message.Text(str)}
}
