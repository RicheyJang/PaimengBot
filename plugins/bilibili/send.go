package bilibili

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/push"

	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var once sync.Once

func startPolling() {
	once.Do(func() {
		_, err := proxy.AddScheduleEveryFunc("30s", checkSubscriptionStatus)
		if err != nil {
			log.Errorf("初始化B站订阅轮询任务失败：%v", err)
		} else {
			log.Infof("成功启动B站订阅轮询任务")
		}
	})
}

func checkSubscriptionStatus() {
	subs := AllSubscription()
	if len(subs) > 1 { // 多个订阅，则打乱顺序
		rand.Shuffle(len(subs), func(i, j int) {
			subs[i], subs[j] = subs[j], subs[i]
		})
	}
	// 检查更新
	var msgs []message.Message
	for _, sub := range subs {
		time.Sleep(time.Second) // 间隔一秒调用
		switch sub.SubType {
		case SubTypeBangumi:
			msgs = checkBangumiStatus(sub)
		case SubTypeUp:
			msgs = checkUpStatus(sub)
		case SubTypeLive:
			msgs = checkLiveStatus(sub)
		}
		// 推送
		if len(msgs) > 0 {
			if proxy.GetConfigBool("atAll") {
				msgs[0] = append(message.Message{message.At(0)}, msgs[0]...)
			}
			f, g := sub.GetFriendsGroups()
			for _, msg := range msgs {
				if len(msg) == 0 {
					continue
				}
				push.Send(push.Target{
					Msg:     msg,
					Friends: f,
					Groups:  g,
				})
			}
		}

	}
}

func checkBangumiStatus(sub Subscription) (msg []message.Message) {
	b, err := NewBangumi().ByMDID(sub.BID)
	if err != nil {
		log.Warnf("获取B站番剧(%v)信息失败：%v", sub.BID, err)
		return nil
	}
	if b.NewEP.Name != sub.BangumiLastIndex { // 更新了
		// 生成消息
		str := fmt.Sprintf("番剧「%v」更新了！\n最新一集：%v", b.Title, b.NewEP.Name)
		link := fmt.Sprintf("\n快去看：https://bilibili.com/bangumi/play/ep%v", b.NewEP.ID)
		if proxy.GetConfigBool("link") {
			str += link
		}
		msg = []message.Message{{message.Text(str)}}
		// 更新状态
		sub.BangumiLastIndex = b.NewEP.Name
		err = UpdateSubsStatus(sub)
		if err != nil {
			log.Warnf("[SQL] 更新B站番剧(%v)订阅信息失败：%v", sub.BID, err)
			return nil // 为防止数据库出错导致的重复推送，直接不推送
		}
	}
	return
}

func checkLiveStatus(sub Subscription) (msg []message.Message) {
	l, err := NewLiveRoom(sub.BID).Info()
	if err != nil {
		log.Warnf("获取B站直播(%v)信息失败：%v", sub.BID, err)
		return nil
	}
	if l.IsOpen() != sub.LiveStatus { // 直播间状态改变
		if l.IsOpen() { // 开播
			// 生成消息
			id := l.ShortID
			if id <= 0 { // 没有直播间短ID
				id = sub.BID
			}
			str := fmt.Sprintf("%v的直播间(%v)开播了！\n标题：%v", l.Anchor.Name, id, l.Title)
			link := fmt.Sprintf("\n快去围观：https://live.bilibili.com/%v", l.ID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			msg = []message.Message{{message.Text(str)}}
		}
		// 更新状态
		sub.LiveStatus = l.IsOpen()
		err = UpdateSubsStatus(sub)
		if err != nil {
			log.Warnf("[SQL] 更新B站直播(%v)信息失败：%v", sub.BID, err)
			return nil // 为防止数据库出错导致的重复推送，直接不推送
		}
	}
	return
}

func checkUpStatus(sub Subscription) (msg []message.Message) {
	ds, _, err := NewUser(sub.BID).Dynamics(0, false)
	if err != nil {
		log.Warnf("获取B站UP主(%v)动态失败：%v", sub.BID, err)
		return nil
	}
	if len(ds) == 0 {
		return nil
	}
	// 取最新的一条
	d := ds[0]
	if d.Time.After(sub.DynamicLastTime) { // 更新了新动态
		var str, link string
		var m []message.MessageSegment

		switch d.Type {
		case DynamicTypeShare: //处理分享动态
			str = fmt.Sprintf("UP主「%v」发表了新动态！", d.Uname)
			link = fmt.Sprintf("\n主页：https://space.bilibili.com/%v/dynamic\n", sub.BID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			m = DynamicTypeShareMessage(d, append(m, message.Text(str)))
		case DynamicTypePic: //
			str = fmt.Sprintf("UP主「%v」发表了新动态！", d.Uname)
			link = fmt.Sprintf("\n主页：https://space.bilibili.com/%v/dynamic\n", sub.BID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			m = DynamicTypePicMessage(d, append(m, message.Text(str)))
		case DynamicTypeText:
			str = fmt.Sprintf("UP主「%v」发表了新动态！", d.Uname)
			link = fmt.Sprintf("\n主页：https://space.bilibili.com/%v/dynamic\n", sub.BID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			m = DynamicTypeTextMessage(d, append(m, message.Text(str)))

		case DynamicTypeVideo: //处理视频动态
			str = fmt.Sprintf("UP主「%v」发布了新视频！", d.Uname)
			title := d.VideoTitle()
			if len(title) > 0 {
				str += "\n标题：" + title
			}
			if len(d.BVID) > 0 {
				link = fmt.Sprintf("\n直链：https://www.bilibili.com/video/%v", d.BVID)
			}
			if proxy.GetConfigBool("link") {
				str += link
			}
			m = append(m, message.Text(str))
		case DynamicTypeRead:
			str = fmt.Sprintf("UP主「%v」发表了新专栏！", d.Uname)
			link = fmt.Sprintf("\n主页：https://space.bilibili.com/%v/dynamic\n", sub.BID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			m = DynamicTypeReadMessage(d, append(m, message.Text(str)))
		}

		msg = []message.Message{m}

		// 更新状态
		sub.DynamicLastTime = d.Time
		err = UpdateSubsStatus(sub)
		if err != nil {
			log.Warnf("[SQL] 更新B站UP主(%v)信息失败：%v", sub.BID, err)
			return nil // 为防止数据库出错导致的重复推送，直接不推送
		}
	}
	return
}
