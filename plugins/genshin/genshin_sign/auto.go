package genshin_sign

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/RicheyJang/PaimengBot/utils/push"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var taskID cron.EntryID

func configReload(fsnotify.Event) error {
	proxy.DeleteSchedule(taskID)
	id, err := proxy.AddScheduleDailyFunc(
		int(proxy.GetConfigInt64("daily.hour")),
		int(proxy.GetConfigInt64("daily.min")),
		autoSignTask)
	if err == nil {
		taskID = id
	}
	return err
}

// 获取所有需要定时签到的用户信息映射
func initCornTasks() map[string]UserInfo {
	iter := proxy.GetLevelDB().NewIterator(levelutil.BytesPrefix([]byte("genshin_")), nil)
	users := make(map[string]UserInfo)
	for iter.Next() {
		keyStr := string(iter.Key())
		value := iter.Value()
		initCookie(keyStr, value, users)
		initUin(keyStr, value, users)
		initEvent(keyStr, value, users)
	}
	iter.Release()
	return users
}

func initCookie(key string, value []byte, users map[string]UserInfo) {
	if strings.HasPrefix(key, "genshin_cookie.u") {
		strValue := ""
		_ = json.Unmarshal(value, &strValue)
		name := key[len("genshin_cookie.u"):]
		userInfo, ok := users[name]
		if ok {
			userInfo.ID = name
			userInfo.cookie = strValue
			users[name] = userInfo
		} else {
			users[name] = UserInfo{ID: name, cookie: strValue}
		}
	}
}

func initUin(key string, value []byte, users map[string]UserInfo) {
	if strings.HasPrefix(key, "genshin_uid.u") {
		strValue := ""
		_ = json.Unmarshal(value, &strValue)
		name := key[len("genshin_uid.u"):]
		userInfo, ok := users[name]
		if ok {
			userInfo.ID = name
			userInfo.Uin = strValue
			users[name] = userInfo
		} else {
			users[name] = UserInfo{ID: name, Uin: strValue}
		}
	}
}

func initEvent(key string, value []byte, users map[string]UserInfo) {
	if strings.HasPrefix(key, "genshin_eventfrom.u") {
		var eventInfo EventFrom
		_ = json.Unmarshal(value, &eventInfo)
		name := key[len("genshin_eventfrom.u"):]
		userInfo, ok := users[name]
		if ok {
			userInfo.ID = name
			userInfo.EventFrom = eventInfo
			users[name] = userInfo
		} else {
			users[name] = UserInfo{ID: name, EventFrom: eventInfo}
		}
	}
}

func autoSignTask() {
	users := initCornTasks()
	for k, user := range users {
		if !user.EventFrom.Auto || len(user.Uin) <= 5 || len(user.cookie) <= 10 {
			continue
		}
		// 执行定时签到
		msg, err := Sign(user.Uin, user.cookie)
		if err != nil {
			log.Warnf("Auto Sign(id=%v, uid=%v) err: %v", user.ID, user.Uin, err)
			time.Sleep(time.Second)
			continue
		}
		log.Infof("米游社自动签到成功：%v", msg)
		// 推送消息
		if user.EventFrom.IsFromGroup && proxy.GetConfigBool("group") { // 来自群的定时签到 且允许推送
			groupID, _ := strconv.ParseInt(user.EventFrom.FromId, 10, 64)
			qq, _ := strconv.ParseInt(k, 10, 64)
			push.Send(push.Target{
				Msg:    message.Message{message.At(qq), images.GenStringMsg(msg)},
				Groups: []int64{groupID},
			})
		} else if !user.EventFrom.IsFromGroup { // 来自个人的定时签到
			qq, _ := strconv.ParseInt(k, 10, 64)
			push.Send(push.Target{
				Msg:     message.Message{images.GenStringMsg(msg)},
				Friends: []int64{qq},
			})
		}
		time.Sleep(2 * time.Second)
	}
}
