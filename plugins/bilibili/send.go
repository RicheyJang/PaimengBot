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
	var msg message.Message
	subs := AllSubscription()
	if len(subs) > 1 { // 多个订阅，则打乱顺序
		rand.Shuffle(len(subs), func(i, j int) {
			subs[i], subs[j] = subs[j], subs[i]
		})
	}
	// 检查更新
	for _, sub := range subs {
		time.Sleep(time.Second) // 间隔一秒调用
		switch sub.SubType {
		case SubTypeBangumi:
			msg = checkBangumiStatus(sub)
		case SubTypeUp:
			msg = checkUpStatus(sub)
		case SubTypeLive:
			msg = checkLiveStatus(sub)
		}
		// 推送
		if len(msg) > 0 {
			f, g := sub.GetFriendsGroups()
			push.Send(push.Target{
				Msg:     msg,
				Friends: f,
				Groups:  g,
			})
		}
	}
}

func checkBangumiStatus(sub Subscription) (msg message.Message) {
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
		msg = message.Message{message.Text(str)}
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

func checkLiveStatus(sub Subscription) (msg message.Message) {
	l, err := NewLiveRoom(sub.BID).Info()
	if err != nil {
		log.Warnf("获取B站直播(%v)信息失败：%v", sub.BID, err)
		return nil
	}
	if l.IsOpen() != sub.LiveStatus { // 直播间状态改变
		if l.IsOpen() { // 开播
			// 生成消息
			str := fmt.Sprintf("%v的直播间(%v)开播了！\n标题：%v", l.Anchor.Name, l.ShortID, l.Title)
			link := fmt.Sprintf("\n快去围观：https://live.bilibili.com/%v", l.ID)
			if proxy.GetConfigBool("link") {
				str += link
			}
			msg = message.Message{message.Text(str)}
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

func checkUpStatus(sub Subscription) (msg message.Message) {
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
		if d.IsVideo() { // 新视频
			str = fmt.Sprintf("UP主「%v」发布了新视频！", d.Uname)
			title := d.VideoTitle()
			if len(title) > 0 {
				str += "\n标题：" + title
			}
			if len(d.BVID) > 0 {
				link = fmt.Sprintf("\n直链：https://www.bilibili.com/video/%v", d.BVID)
			}
		} else { // 普通动态
			str = fmt.Sprintf("UP主「%v」发表了新动态！", d.Uname)
			link = fmt.Sprintf("\n主页：https://space.bilibili.com/%v/dynamic", sub.BID)
		}
		if proxy.GetConfigBool("link") {
			str += link
		}
		msg = message.Message{message.Text(str)}
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
