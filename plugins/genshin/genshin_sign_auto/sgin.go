package genshin_sign_auto

import (
	"encoding/json"
	"fmt"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_public"
	"github.com/RicheyJang/PaimengBot/plugins/genshin/genshin_sign"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"strings"
	"time"
)

var info = manager.PluginInfo{
	Name: "定时签到",
	Usage: `如果你填写了对应的cookie
将会自动在查询对应的信息 说 定时签到 打开/关闭 就可以啦
` + genshin_public.GetInitializaationPrompt(),
	Classify: "原神相关",
}
var proxy *manager.PluginProxy

type EventFrom struct {
	isFromGroup bool
	fromId      string
	qq          string
	auto        bool
}
type UserInfo struct {
	ID         string
	uin        string
	cookie     string
	event_from EventFrom
}

var users = map[string]UserInfo{}

func init() {
	proxy = manager.RegisterPlugin(info) // [3] 使用插件信息初始化插件代理
	if proxy == nil {                    // 若初始化失败，请return，失败原因会在日志中打印
		return
	}
	// [4] 此处进行其它初始化操作
	// 添加定时签到任务
	proxy.AddScheduleDailyFunc(14, 00, auto_sign)
	proxy.OnCommands([]string{"自动签到", "定时签到"}).SetBlock(true).SetPriority(3).Handle(sign)
}
func init_corn_taks() {

	db := proxy.GetLevelDB()
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		key_str := string(key)
		value := iter.Value()
		value_str := string(value)

		init_cookie(key_str, value_str)
		init_uin(key_str, value_str)
		init_event(key_str, value)
		//fmt.Println(fmt.Sprintf("%s|%s",key,value))
	}
	iter.Release()
	//err = iter.Error()
	//fmt.Println(users)
}

func init_cookie(key string, value string) {
	index := strings.Index(key, "genshin_cookie.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_cookie.u"):]
		user_info, ok := users[name]
		fmt.Println(name)
		if ok {
			user_info.ID = name
			user_info.cookie = value
		} else {
			userInfo := UserInfo{name, "", value, EventFrom{false, "", "", false}}
			users[name] = userInfo
		}
	}
}

func init_uin(key string, value string) {
	index := strings.Index(key, "genshin_uid.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_uid.u"):]
		user_info, ok := users[name]
		fmt.Println(name)
		if ok {
			user_info.ID = name
			user_info.uin = value
		} else {
			userInfo := UserInfo{name, value, "", EventFrom{false, "", "", false}}
			users[name] = userInfo
		}
	}
}
func init_event(key string, value []byte) {
	index := strings.Index(key, "genshin_eventfrom.u")
	if index != -1 {
		// cookie
		name := key[index+len("genshin_eventfrom.u"):]
		user_info, ok := users[name]
		fmt.Println(name)
		if ok {
			_ = json.Unmarshal(value, &user_info.event_from)
			user_info.ID = name
		} else {
			event_from := EventFrom{}
			_ = json.Unmarshal(value, &event_from)
			userInfo := UserInfo{name, "", "", event_from}
			users[name] = userInfo
		}
	}
}

func auto_sign() {
	init_corn_taks()
	for k, v := range users {
		if v.event_from.auto {
			// 执行定时任务
			ctx := utils.GetBotCtx()
			if v.event_from.isFromGroup {
				// 来自群的定时
				msg, err := genshin_sign.Sign(v.uin, v.cookie)
				if err != nil {
					msg = "定时任务执行失败\n" + err.Error()
				} else {
					msg = "定时任务执行完成:\n" + msg
				}
				group_id, _ := strconv.ParseInt(v.event_from.fromId, 10, 64)
				qq_id, _ := strconv.ParseInt(k, 10, 64)
				ctx.SendGroupMessage(group_id, message.Message{message.At(qq_id), message.Text(msg)})
			} else {
				// 来自个人的定时
				qq_id, _ := strconv.ParseInt(k, 10, 64)
				msg, err := genshin_sign.Sign(v.uin, v.cookie)
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

func sign(ctx *zero.Ctx) {
	_, _, cookie_msg, err := genshin_public.GetUidCookieById(ctx.Event.UserID)
	if err != nil {
		ctx.Send(images.GenStringMsg(cookie_msg))
		return
	}
	// 接收参数 判断是开还是关
	args := utils.GetArgs(ctx)
	if isIn(args, "开") {
		// 添加定时
		if ctx.Event.GroupID != 0 {
			//来自群聊
			PutEventFrom(ctx.Event.UserID, EventFrom{
				true,
				strconv.FormatInt(ctx.Event.GroupID, 10),
				strconv.FormatInt(ctx.Event.UserID, 10),
				true})
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
		ctx.Send(`你在做什么？
	该功能的使用方法是：
		自动签到 开/关
		定时签到 开/关
	这样，机器人每天就会定时帮你签到了，还会在你打开该功能的地方告诉你`)
		return
	}
	return
	//
	//msg, err := genshin_sign.Sign(user_uid, user_cookie)
	//if err != nil {
	//	ctx.Send(images.GenStringMsg(msg))
	//}
	//ctx.Send(message.Text(fmt.Sprintf("签到:%s", msg)))
}

func isIn(str string, deps string) bool {
	index := strings.Index(str, deps)
	if index != -1 {
		return true
	} else {
		return false
	}
}

func GetEventFrom(id int64) (EventFrom, error) {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	v, err := proxy.GetLevelDB().Get([]byte(key), nil)
	event_from := EventFrom{false, "", "", false}
	if err != nil {
		return event_from, err
	}

	_ = json.Unmarshal(v, &event_from)
	return event_from, err
}

func PutEventFrom(id int64, u EventFrom) error {
	key := fmt.Sprintf("genshin_eventfrom.u%v", id)
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return proxy.GetLevelDB().Put([]byte(key), value, nil)
}
