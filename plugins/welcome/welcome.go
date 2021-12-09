package welcome

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	"github.com/RicheyJang/PaimengBot/utils/consts"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm/clause"
)

var info = manager.PluginInfo{
	Name: "群欢迎消息",
	Usage: `用法：
	设置群欢迎消息 [消息...]：当有新人加群时，自动发送所设置的欢迎消息+@新人`,
	AdminLevel: 3,
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnCommands([]string{"设置群欢迎消息", "自定义群欢迎消息"}, zero.OnlyGroup).
		SetBlock(true).SetPriority(3).Handle(setGroupWelcome)
	proxy.OnNotice(utils.CheckDetailType("group_increase"), func(ctx *zero.Ctx) bool {
		return ctx.Event.SelfID != ctx.Event.UserID
	}).SetBlock(false).SecondPriority().Handle(handleIncrease)
}

func setGroupWelcome(ctx *zero.Ctx) {
	var welmsg message.Message
	// 消除首段消息前的Bot昵称
	first := ctx.Event.Message[0].Data["text"]
	first = strings.TrimLeft(first, " \t") // Trim!
	for _, nickname := range utils.GetBotConfig().NickName {
		if strings.HasPrefix(first, nickname) {
			first = first[len(nickname):]
			break
		}
	}
	// 消除首段消息前的命令
	first = strings.Replace(first, utils.GetCommand(ctx), "", 1)
	first = strings.Trim(first, " \t") // Trim!
	// 拼接消息
	if len(first) > 0 {
		welmsg = append(welmsg, message.Text(first))
	}
	if len(ctx.Event.Message) > 1 {
		for i, msg := range ctx.Event.Message[1:] {
			if msg.Type == "image" { // 将收到的图片URL存至本地
				msg = recvImage2Local(ctx.Event.GroupID, int64(i), msg)
			}
			welmsg = append(welmsg, msg)
		}
	}
	if len(welmsg) == 0 { // 欢迎消息最终为空
		ctx.Send("欢迎词呢？")
		return
	}
	welstr := utils.JsonString(welmsg)
	// 更新数据库
	preGroup := dao.GroupSetting{
		ID:      ctx.Event.GroupID,
		Welcome: welstr,
	}
	if err := proxy.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"welcome"}), // Upsert
	}).Create(&preGroup).Error; err != nil {
		log.Errorf("set group(%v) welcome error(sql): %v", ctx.Event.GroupID, err)
		ctx.Send("设置失败了...")
		return
	}
	log.Infof("群%v的欢迎消息设置为：%v", preGroup.ID, preGroup.Welcome)
	// 发送测试 ----
	// 将欢迎消息中的图片（本地）转换为可发送格式
	var sendMsg message.Message
	for _, seg := range welmsg {
		if seg.Type == "image" {
			seg = localImage2Send(seg)
		}
		sendMsg = append(sendMsg, seg)
	}
	// 发送
	ctx.Send(message.Text("好哒，新人入群欢迎消息设置为：\n").String() + message.At(ctx.Event.UserID).String() + sendMsg.String())
}

// 有新人入群时
func handleIncrease(ctx *zero.Ctx) {
	var groupS dao.GroupSetting
	res := proxy.GetDB().Select("id", "welcome").Take(&groupS, ctx.Event.GroupID)
	if res.RowsAffected == 0 || len(groupS.Welcome) == 0 {
		return
	}
	// 将欢迎消息中的图片（本地）转换为可发送格式
	msg := message.ParseMessage([]byte(groupS.Welcome))
	var sendMsg message.Message
	for _, seg := range msg {
		if seg.Type == "image" {
			seg = localImage2Send(seg)
		}
		sendMsg = append(sendMsg, seg)
	}
	// 发送
	ctx.SendGroupMessage(ctx.Event.GroupID, message.At(ctx.Event.UserID).String()+sendMsg.String())
}

// 收到的图片消息，存储至本地消息
func recvImage2Local(groupID, num int64, msg message.MessageSegment) message.MessageSegment {
	url := utils.GetImageURL(msg)
	if len(url) == 0 {
		return msg
	}
	if _, err := utils.MakeDir(consts.GroupImageDir); err != nil {
		log.Warnf("recvImage2Local mkdir err: %v", err)
		return msg
	}
	filename := utils.PathJoin(consts.GroupImageDir, fmt.Sprintf("%v-%v.jpg", groupID, num))
	if err := client.DownloadToFile(filename, url, 2); err != nil {
		log.Warnf("recvImage2Local err: %v", err)
		return msg
	}
	abs, err := filepath.Abs(filename)
	if err != nil {
		log.Warnf("recvImage2Local file abs err: %v", err)
		return msg
	}
	return message.Image("file:///" + abs) // 本地绝对路径图片
}

// 本地的图片消息，自动转换成可发送消息（若OneBot收发端不在本地，改用Base64）
func localImage2Send(msg message.MessageSegment) message.MessageSegment {
	if utils.IsOneBotLocal() || !strings.HasPrefix(msg.Data["file"], "file:///") {
		return msg
	}
	// 打开文件
	filename := strings.Replace(msg.Data["file"], "file:///", "", 1)
	if len(filename) == 0 {
		log.Warnf("localImage2Send filename is empty")
		return msg
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Warnf("localImage2Send open file(%v) err: %v", filename, err)
		return msg
	}
	defer f.Close()
	// Base64
	resultBuff := bytes.NewBuffer(nil) // 结果缓冲区
	// 新建Base64编码器（Base64结果写入结果缓冲区resultBuff）
	encoder := base64.NewEncoder(base64.StdEncoding, resultBuff)
	// 将文件写入Base64编码器
	_, err = io.Copy(encoder, f)
	if err != nil {
		log.Warnf("localImage2Send base64 copy err: %v", err)
		_ = encoder.Close()
		return msg
	}
	// 结束Base64编码
	err = encoder.Close()
	if err != nil {
		log.Warnf("localImage2Send base64 encode err: %v", err)
		return msg
	}
	return message.Image("base64://" + resultBuff.String())
}
