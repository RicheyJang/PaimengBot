package sc

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/basic/dao"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type signInfo struct {
	id       int64
	name     string
	double   bool
	orgFavor float64
	addFavor float64
	orgCoin  float64
	addCoin  float64
	signDays int
	lastSign time.Time // 最近一次签到时间
}

func (s signInfo) genMessage() message.Message {
	W, H, avaSize := 500, 300, 100
	img := images.NewImageCtxWithBGColor(W, H, "white")
	// 背景 linear-gradient(135deg,#fff5c3,#9452a5)
	gra := gg.NewLinearGradient(0, 0, float64(W), float64(H))
	gra.AddColorStop(0, color.NRGBA{R: uint8(255), G: uint8(245), B: uint8(195), A: uint8(200)})
	gra.AddColorStop(1, color.NRGBA{R: uint8(148), G: uint8(82), B: uint8(165), A: uint8(200)})
	img.Push()
	img.DrawRectangle(0, 0, float64(W), float64(H))
	img.SetFillStyle(gra)
	img.Fill()
	img.Pop()
	// 写昵称+ID
	str := fmt.Sprintf("%s(%d)", strings.TrimSpace(s.name), s.id)
	err := img.PasteStringDefault(str, 24, 1.3, 40, 20, float64(W))
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	// 画头像
	height := 70
	avaReader, err := utils.GetQQAvatar(s.id, avaSize)
	if err != nil {
		log.Warnf("GetQQAvatar err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	ava, _, err := image.Decode(avaReader)
	_ = avaReader.Close()
	if err != nil {
		log.Warnf("Avatar Decode err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	ava = images.ClipImgToCircle(ava)
	img.DrawImage(ava, 20, height)
	// 头像旁边的文字
	level, up := LevelAt(s.orgFavor + s.addFavor)
	err = img.PasteStringDefault(fmt.Sprintf("连续签到%d天\nLv%d", s.signDays, level),
		18, 1.88, float64(avaSize+30), float64(height), float64(W))
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	// 头像旁边的等级进度条
	img.SetHexColor("#6eb7f0")
	length := 290.0
	img.DrawRoundedRectangle(float64(20+avaSize+60), float64(height+30), length, 35, 5)
	img.Stroke()
	length *= 1 - up/SumFavorAt(level)
	img.DrawRoundedRectangle(float64(20+avaSize+60), float64(height+30), length, 35, 5)
	img.Fill()
	// 头像旁边的还需多少升级的文字
	err = img.PasteStringDefault(fmt.Sprintf("总好感度%.2f, 还需%.2f升级", s.orgFavor+s.addFavor, up),
		18, 1, float64(avaSize+30), float64(height+70), float64(W))
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	// 今日成果文字
	height += avaSize + 30
	err = img.PasteStringDefault(fmt.Sprintf("今日好感度 + %.2f\n今日获得 %.0f%s", s.addFavor, RealCoin(s.addCoin), Unit()),
		26, 1.62, 30, float64(height), float64(W))
	if err != nil {
		log.Warnf("PasteStringDefault err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	// 双倍
	if s.double {
		img.SetHexColor("#ffd591")
		img.DrawCircle(float64(W-75), float64(height+35), 40)
		img.Fill()
		if err = img.UseDefaultFont(24); err != nil {
			log.Warnf("UseDefaultFont err: %v", err)
			return message.Message{message.Text("签到成功\n" + s.String())}
		}
		img.SetHexColor("#ff4d4f")
		img.DrawString("双倍", float64(W-100), float64(height+40))
	}
	// 生成消息
	imgMsg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("GenMessageAuto err: %v", err)
		return message.Message{message.Text("签到成功\n" + s.String())}
	}
	return message.Message{imgMsg}
}

func (s signInfo) String() string {
	var doubleStr string
	if s.double {
		doubleStr = "✪ ω ✪ 双倍！\n"
	}
	return fmt.Sprintf("%s已连续签到%v天\n好感度：%.2f(+%.2f)\n财富：%.0f(+%.0f)%s",
		doubleStr,
		s.signDays,
		s.orgFavor+s.addFavor, s.addFavor,
		RealCoin(s.orgCoin+s.addCoin), RealCoin(s.addCoin), Unit())
}

func genRankMessage(ctx *zero.Ctx, users []dao.UserOwn, key string) (msg message.MessageSegment, err error) {
	var values []images.UserValue
	if key == "favor" { // 好感度
		for _, user := range users {
			values = append(values, images.UserValue{
				ID:      user.ID,
				Value:   user.Favor,
				FmtPrec: 2,
			})
		}
		return images.GenQQRankMsgWithValue("好感度排行榜", values, "")
	} else { // 财富
		for _, user := range users {
			values = append(values, images.UserValue{
				ID:    user.ID,
				Value: RealCoin(user.Wealth),
			})
		}
		return images.GenQQRankMsgWithValue("财富排行榜", values, Unit())
	}
}
