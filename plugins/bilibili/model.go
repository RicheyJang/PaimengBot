package bilibili

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"

	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
)

const (
	SubTypeUp      = "up"
	SubTypeBangumi = "bangumi"
	SubTypeLive    = "live"

	SubUserAll = "all"
)

func init() {
	err := manager.GetDB().AutoMigrate(&Subscription{})
	if err != nil {
		log.Errorf("[SQL] Bilibili Subscription初始化失败，err: %v", err)
	}
}

type Subscription struct {
	ID       int
	SubUsers string // 格式：若为私人，则直接是ID；若为群订阅，则为 群ID:发起用户ID；多个User用,分隔
	SubType  string `gorm:"uniqueIndex:idx_sub_item;size:64"`
	BID      int64  `gorm:"uniqueIndex:idx_sub_item"` // UP主ID、番剧ID、直播间ID
	// 状态相关：
	BangumiLastIndex string    // 番剧：最后一集Index
	DynamicLastTime  time.Time // UP动态：最后一条动态时间戳
	LiveStatus       bool      // 直播间：开播状态
}

func (s *Subscription) TableName() string {
	return "t_bilibili_subscriptions"
}

func (s Subscription) GenMessage(showUsers bool) message.Message {
	str := "订阅ID：" + strconv.Itoa(s.ID) + "\n"
	switch s.SubType {
	case SubTypeBangumi:
		b, err := NewBangumi().ByMDID(s.BID)
		if err != nil {
			log.Errorf("Get Bangumi by MDID error: %v", err)
			str += "番剧：ID=" + strconv.FormatInt(s.BID, 10)
		} else {
			str += "番剧：" + b.Title + "\n" + "最新一集：" + b.NewEP.Name
		}
	case SubTypeUp:
		u, err := NewUser(s.BID).Info()
		if err != nil {
			log.Errorf("Get User by MID error: %v", err)
			str += "UP主：ID=" + strconv.FormatInt(s.BID, 10)
		} else {
			str += "UP主：" + u.Name
		}
	case SubTypeLive:
		l, err := NewLiveRoom(s.BID).Info()
		if err != nil {
			log.Errorf("Get LiveRoom by roomID error: %v", err)
			str += "直播间：ID=" + strconv.FormatInt(s.BID, 10)
		} else {
			str += l.Anchor.Name + "的直播间：ID=" + strconv.FormatInt(s.BID, 10)
		}
	default: // 其他类型，不做处理
		return nil
	}
	if showUsers {
		str += "\n" + "订阅用户：" + s.GenUsersText()
	}
	return message.Message{message.Text(str)}
}

func (s Subscription) GenUsersText() string {
	users := strings.Split(s.SubUsers, ",")
	var str string
	for _, user := range users {
		if len(str) > 0 {
			str += "、"
		}
		if index := strings.Index(user, ":"); index > 0 {
			str += "群" + user[:index] + "(发起订阅者:" + user[index+1:] + ")"
		} else {
			str += user
		}
	}
	return str
}

func (s Subscription) GetFriendsGroups() (friends, groups []int64) {
	users := strings.Split(s.SubUsers, ",")
	for _, user := range users {
		if index := strings.Index(user, ":"); index > 0 {
			group, err := strconv.ParseInt(user[:index], 10, 64)
			if err != nil {
				continue
			}
			groups = append(groups, group)
		} else {
			friend, err := strconv.ParseInt(user, 10, 64)
			if err != nil {
				continue
			}
			friends = append(friends, friend)
		}
	}
	return
}

// AllSubscription 获取所有订阅
func AllSubscription() []Subscription {
	var subs []Subscription
	err := proxy.GetDB().Find(&subs).Error
	if err != nil {
		log.Errorf("Get all bilibili subscription error: %v", err)
	}
	return subs
}

// GetSubForPrimary 获取私人订阅列表
func GetSubForPrimary(userID int64) []Subscription {
	var subs []Subscription
	idFormat := strconv.FormatInt(userID, 10)
	err := proxy.GetDB().Where("sub_users LIKE ?", "%"+idFormat+"%").Find(&subs).Error
	if err != nil {
		log.Errorf("Get bilibili subscription for primary user error: %v", err)
		return nil
	}
	// 再次检查
	var rsp []Subscription
	for _, sub := range subs {
		users := strings.Split(sub.SubUsers, ",")
		if utils.StringSliceContain(users, idFormat) {
			rsp = append(rsp, sub)
		}
	}
	return rsp
}

// GetSubForGroup 获取群订阅列表
func GetSubForGroup(groupID int64) []Subscription {
	var subs []Subscription
	idFormat := strconv.FormatInt(groupID, 10) + ":"
	err := proxy.GetDB().Where("sub_users LIKE ?", "%"+idFormat+"%").Find(&subs).Error
	if err != nil {
		log.Errorf("Get bilibili subscription for group error: %v", err)
		return nil
	}
	// 再次检查
	var rsp []Subscription
	for _, sub := range subs {
		users := strings.Split(sub.SubUsers, ",")
		for _, user := range users {
			if strings.HasPrefix(user, idFormat) {
				rsp = append(rsp, sub)
				break
			}
		}
	}
	return rsp
}

// AddSubscription 添加订阅
func AddSubscription(sub Subscription) error {
	err := proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		oldSub := Subscription{}
		result := tx.Where(&Subscription{
			SubType: sub.SubType,
			BID:     sub.BID,
		}).First(&oldSub)
		if result.RowsAffected == 0 { // 新增
			sub.ID = 0
			sub.DynamicLastTime = time.Now()
			result = tx.Create(&sub)
			if result.Error != nil {
				return fmt.Errorf("add bilibili subscription error: %v", result.Error)
			}
		} else { // 更新用户
			newUsers := utils.MergeStringSlices(strings.Split(sub.SubUsers, ","),
				strings.Split(oldSub.SubUsers, ","))
			result = tx.Model(&oldSub).Update("sub_users", strings.Join(newUsers, ","))
			if result.Error != nil {
				return fmt.Errorf("update bilibili subscription error: %v", result.Error)
			}
		}
		return nil
	})
	if err == nil {
		startPolling()
	}
	return err
}

func UpdateSubsStatus(sub Subscription) error {
	copySub := sub
	return proxy.GetDB().Model(&copySub).
		Select("bangumi_last_index", "dynamic_last_time", "live_status").Updates(sub).Error
}

// DeleteSubscription 删除订阅，必须指定SubUsers(=SubUserAll则该订阅全部删除)，另外至少需要SubType和BID，或单独指定ID
func DeleteSubscription(sub Subscription) error {
	return proxy.GetDB().Transaction(func(tx *gorm.DB) error {
		oldSub := Subscription{}
		result := tx.Where(&Subscription{
			ID:      sub.ID,
			SubType: sub.SubType,
			BID:     sub.BID,
		}).First(&oldSub)
		if result.Error != nil {
			return fmt.Errorf("before delete: get bilibili subscription error: %v", result.Error)
		}
		if result.RowsAffected == 0 { // 不存在
			return nil
		}
		// 筛选需要删除的用户
		var newUsers []string
		delUsers := strings.Split(sub.SubUsers, ",")
		oldUsers := strings.Split(oldSub.SubUsers, ",")
		for _, old := range oldUsers {
			couldAdd := true
			for _, del := range delUsers {
				// ID一致，或群ID前缀一致
				if old == del ||
					(strings.ContainsRune(old, ':') && strings.HasSuffix(del, ":") && strings.HasPrefix(old, del)) {
					couldAdd = false
					break
				}
			}
			if couldAdd {
				newUsers = append(newUsers, old)
			}
		}
		// 执行删除
		if len(newUsers) == 0 || sub.SubUsers == SubUserAll { // 没有其它用户订阅了或指令删除全部
			return tx.Delete(&oldSub).Error
		} else {
			return tx.Model(&oldSub).Update("sub_users", strings.Join(newUsers, ",")).Error
		}
	})
}
