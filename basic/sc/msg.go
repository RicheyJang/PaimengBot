package sc

import (
	"fmt"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type signInfo struct {
	id       int64
	double   bool
	orgFavor float64
	addFavor float64
	orgCoin  float64
	addCoin  float64
	signDays int
	lastSign time.Time // 最近一次签到时间
}

func (s signInfo) genMessage() message.Message {
	// TODO 绘图
	return message.Message{message.Text("签到成功\n" + s.String())}
}

func (s signInfo) String() string {
	var doubleStr string
	if s.double {
		doubleStr = "✪ ω ✪ 双倍！\n"
	}
	return fmt.Sprintf("%s已连续签到%v天\n好感度：%.2f(+%.2f)\n财富：%.0f(+%.0f)%s",
		doubleStr,
		s.signDays,
		s.orgFavor+s.addFavor, s.addFavor,
		RealCoin(s.orgCoin+s.addCoin), RealCoin(s.addCoin), Unit())
}
