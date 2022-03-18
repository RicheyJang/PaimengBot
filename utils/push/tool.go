package push

import (
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func intersectionWithJsonArray(a []int64, b []gjson.Result, bKey string) []int64 {
	m := make(map[int64]bool)
	for _, v := range a {
		m[v] = true
	}
	var ret []int64
	for _, obj := range b {
		v := obj.Get(bKey).Int()
		if m[v] {
			ret = append(ret, v)
		}
	}
	return ret
}

func (t Target) preprocessPrimaryMsg(id int64, msg message.Message) (res message.Message) {
	for _, m := range msg {
		if m.Type != "at" { // 除@外的消息
			res = append(res, m)
		}
	}
	return
}

func (t Target) preprocessGroupMsg(id int64, msg message.Message) message.Message {
	for i, m := range msg {
		if m.Type == "at" && m.Data["qq"] == "all" { // @全体成员消息
			rsp := t.GetCtx().CallAction("get_group_at_all_remain", zero.Params{
				"group_id": id,
			})
			if rsp.RetCode == 0 &&
				!(rsp.Data.Get("can_at_all").Bool() && rsp.Data.Get("remain_at_all_count_for_uin").Int() > 0) {
				// 无法@全体成员
				log.Infof("在群%d中无法@全体成员或当日次数已全部用完，将去除@全体成员发送", id)
				msg = append(msg[:i], msg[i+1:]...)
				break
			}
		}
	}
	return msg
}
