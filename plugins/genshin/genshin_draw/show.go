package genshin_draw

import (
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func showNowPool(ctx *zero.Ctx) {
	for tp := range poolTypeMap {
		pools := LoadPools(tp)
		for _, pool := range pools {
			msg := displaySinglePool(&pool)
			if len(msg) > 0 {
				ctx.Send(msg)
			}
		}
	}
}

func displaySinglePool(pool *DrawPool) (msg message.Message) {
	if pool.EndTimestamp <= time.Now().Unix() { // 已过时
		return
	}
	msg = append(msg, message.Text(pool.Title+"\n"))
	if len(pool.PicURL) > 0 {
		msg = append(msg, message.Image(pool.PicURL))
	}
	msg = append(msg, message.Text("\n卡池名："+pool.Name))
	if time.Now().AddDate(0, 2, 0).After(time.Unix(pool.EndTimestamp, 0)) {
		tm := time.Unix(pool.EndTimestamp, 0).Format("2006-01-02 15:04")
		msg = append(msg, message.Text("\n结束时间："+tm))
	}
	if len(pool.Limit5) > 0 {
		msg = append(msg, message.Text("\nUP 5★："+strings.Join(pool.Limit5, "、")))
	}
	if len(pool.Limit4) > 0 {
		msg = append(msg, message.Text("\nUP 4★："+strings.Join(pool.Limit4, "、")))
	}
	return
}
