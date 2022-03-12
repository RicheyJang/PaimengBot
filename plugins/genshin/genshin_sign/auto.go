package genshin_sign

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/plugins/genshin/mihoyo"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var task_id cron.EntryID

type EventFrom struct {
	IsFromGroup bool
	FromId      string
	Qq          string
	Auto        bool
}
type UserInfo struct {
	ID        string
	Uin       string
	cookie    string
	EventFrom EventFrom
}

func configReload(event fsnotify.Event) error {
	proxy.DeleteSchedule(task_id)
	id, err := proxy.AddScheduleDailyFunc(
		int(proxy.GetConfigInt64("daily.hour")),
		int(proxy.GetConfigInt64("daily.min")),
		auto_sign)
	task_id = id
	return err
}

func querySignHandler(ctx *zero.Ctx) {
	event_from, _ := GetEventFrom(ctx.Event.UserID)
	value, _ := json.Marshal(event_from)
	ctx.Send(message.Text(fmt.Sprintf(string(value))))
	return
}

func init_corn_taks() map[string]UserInfo {

	db := proxy.GetLevelDB()
	iter := db.NewIterator(nil, nil)
	users := map[string]UserInfo{}
	for iter.Next() {
		key := iter.Key()
		key_str := string(key)
		value := iter.Value()
		init_cookie(key_str, value, &users)
		init_uin(key_str, value, &users)
		init_event(key_str, value, &users)
		//fmt.Println(fmt.Sprintf("%s|%s",key,value))
	}
	iter.Release()
	//err = iter.Error()
	//fmt.Println(users)
	return users
}

func init_cookie(key string, value []byte, users *map[string]UserInfo) {
	index := strings.Index(key, "genshin_cookie.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_cookie.u"):]
		user_info, ok := (*users)[name]
		str_value := ""
		_ = json.Unmarshal(value, &str_value)
		if ok {
			user_info.ID = name
			user_info.cookie = str_value
			(*users)[name] = user_info
		} else {
			userInfo := UserInfo{name, "", str_value, EventFrom{false, "", "", false}}
			(*users)[name] = userInfo
		}
	}
}

func init_uin(key string, value []byte, users *map[string]UserInfo) {
	index := strings.Index(key, "genshin_uid.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_uid.u"):]
		user_info, ok := (*users)[name]
		str_value := ""
		_ = json.Unmarshal(value, &str_value)
		if ok {
			user_info.ID = name
			user_info.Uin = str_value
			(*users)[name] = user_info
		} else {
			userInfo := UserInfo{name, str_value, "", EventFrom{false, "", "", false}}
			(*users)[name] = userInfo
		}
	}
}
func init_event(key string, value []byte, users *map[string]UserInfo) {
	index := strings.Index(key, "genshin_eventfrom.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_eventfrom.u"):]
		user_info, ok := (*users)[name]
		event_info := EventFrom{false, "", "", false}
		_ = json.Unmarshal(value, &event_info)
		if ok {
			user_info.EventFrom = event_info
			user_info.ID = name
			(*users)[name] = user_info
		} else {
			userInfo := UserInfo{name, "", "", event_info}
			(*users)[name] = userInfo
		}
	}
}

func auto_sign() {
	users := init_corn_taks()
	for k, v := range users {
		if v.EventFrom.Auto {
			// 执行定时任务
			ctx := utils.GetBotCtx()
			if v.EventFrom.IsFromGroup {
				// 来自群的定时
				msg, err := Sign(v.Uin, v.cookie)
				if err != nil {
					msg = "定时任务执行失败\n" + err.Error()
				} else {
					msg = "定时任务执行完成:\n" + msg
				}
				group_id, _ := strconv.ParseInt(v.EventFrom.FromId, 10, 64)
				qq_id, _ := strconv.ParseInt(k, 10, 64)
				ctx.SendGroupMessage(group_id, message.Message{message.At(qq_id), message.Text(msg)})
			} else {
				// 来自个人的定时
				qq_id, _ := strconv.ParseInt(k, 10, 64)
				msg, err := Sign(v.Uin, v.cookie)
				if err != nil {
					msg = "定时任务执行失败\n" + err.Error()
				} else {
					msg = "定时任务执行完成:\n" + msg
				}
				ctx.SendPrivateMessage(qq_id, message.Text(msg))
			}
			time.Sleep(2 * time.Second)
		}
	}

}

func autoSignHandler(ctx *zero.Ctx) {
	_, _, cookie_msg, err := mihoyo.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookie_msg))
		return
	}
	// 接收参数 判断是开还是关
	args := utils.GetArgs(ctx)
	if isIn(args, "开") == true {
		// 添加定时
		if ctx.Event.GroupID != 0 {
			//来自群聊
			err := PutEventFrom(ctx.Event.UserID, EventFrom{
				true,
				strconv.FormatInt(ctx.Event.GroupID, 10),
				strconv.FormatInt(ctx.Event.UserID, 10),
				true})
			if err != nil {
				fmt.Println(err.Error())
			}
			ctx.Send(message.Message{
				message.At(ctx.Event.UserID),
				message.Text("定时签到已打开"),
			})
		} else {
			PutEventFrom(ctx.Event.UserID,
				EventFrom{false,
					strconv.FormatInt(ctx.Event.UserID, 10),
					strconv.FormatInt(ctx.Event.UserID, 10),
					true})
			ctx.Send(message.Text("定时签到已打开"))
		}
	} else if isIn(args, "关") {
		if ctx.Event.GroupID != 0 {
			//来自群聊
			PutEventFrom(ctx.Event.UserID, EventFrom{
				true,
				strconv.FormatInt(ctx.Event.GroupID, 10),
				strconv.FormatInt(ctx.Event.UserID, 10),
				false})
			ctx.Send(message.Message{
				message.At(ctx.Event.UserID),
				message.Text("定时签到已关闭"),
			})
		} else {
			PutEventFrom(ctx.Event.UserID,
				EventFrom{false,
					strconv.FormatInt(ctx.Event.UserID, 10),
					strconv.FormatInt(ctx.Event.UserID, 10),
					false})
			ctx.Send(message.Text("定时签到已关闭"))
		}
	} else {
		// 不知道啥情况
		ctx.Send(`？可以看看帮助`)
		return
	}
	return
}

func isIn(str string, deps string) bool {
	index := strings.Index(str, deps)
	if index != -1 {
		return true
	} else {
		return false
	}
}

func GetEventFrom(id int64) (event_from EventFrom, e error) {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	if err != nil {
		e = err
		return
	}

	_ = json.Unmarshal(v, &event_from)
	return
}

func PutEventFrom(id int64, u EventFrom) error {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}
