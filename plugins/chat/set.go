package chat

import (
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

// SetDialogue 新增或修改一个问答
func SetDialogue(groupID int64, question string, answer message.Message) error {
	groupD := GroupChatDialogue{
		GroupID:  groupID,
		Question: question,
		Answer:   utils.JsonString(answer),
	}
	return proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "question"}},
		UpdateAll: true,
	}).Create(&groupD).Error
}

// DeleteDialogue 根据问题删除一个问答
func DeleteDialogue(groupID int64, question string) error {
	groupD := GroupChatDialogue{
		GroupID:  groupID,
		Question: question,
	}
	return proxy.GetDB().Delete(&groupD).Error
}

// GetDialogue 根据群号和问题获取answer消息
func GetDialogue(groupID int64, question string) message.Message {
	resD := GroupChatDialogue{}
	rows := proxy.GetDB().Where(&GroupChatDialogue{
		GroupID:  groupID,
		Question: question,
	}, "group_id", "question").Find(&resD).RowsAffected
	if rows == 0 { // 数据库中没有，尝试从问答集文件中读取
		return GetDialogueByFilesRandom(groupID, question)
	}
	return message.ParseMessage([]byte(resD.Answer))
}

// GetAllQuestion 获取指定个群可以触发的所有问答的问题
func GetAllQuestion(groupID int64) []string {
	var resD []GroupChatDialogue
	proxy.GetDB().Where("group_id = ?", groupID).Or("group_id = ?", 0).Find(&resD)
	var qs []string
	for _, r := range resD {
		qs = append(qs, r.Question)
	}
	return utils.MergeStringSlices(qs)
}

type GroupChatDialogue struct {
	GroupID  int64  `gorm:"column:group_id;primaryKey;autoIncrement:false"`
	Question string `gorm:"column:question;primaryKey;autoIncrement:false"`
	Answer   string `gorm:"column:answer"`
}

func init() {
	err := manager.GetDB().AutoMigrate(&GroupChatDialogue{})
	if err != nil {
		log.Errorf("[SQL] GroupChatDialogue 初始化失败, err: %v", err)
	}
}
