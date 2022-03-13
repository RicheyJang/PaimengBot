package sc

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
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
	defer func() {
		if err != nil { // 生成图片失败时，生成文字消息
			log.Warnf("genRankMessage err: %v", err)
			str := "排行榜："
			for _, user := range users {
				if key == "favor" { // 好感度
					str += "\n" + fmt.Sprintf("%v的好感度: %.2f", user.ID, user.Favor)
				} else { // 财富
					str += "\n" + fmt.Sprintf("%v的财富: %.0f%s", user.ID, RealCoin(user.Wealth), Unit())
				}
			}
			msg = message.Text(str)
		}
	}()
	var avaReader io.ReadCloser
	w, avaSize, idSize, lineLength, lineHeight, fontSize, height := 600, 100, 180, 380.0, 50.0, 24.0, 10
	img := images.NewImageCtxWithBGColor(w+avaSize+30, len(users)*(avaSize+20)+30, "white")
	for _, user := range users {
		// 画头像
		avaReader, err = utils.GetQQAvatar(user.ID, avaSize)
		if err != nil {
			return msg, err
		}
		ava, _, err := image.Decode(avaReader)
		_ = avaReader.Close()
		if err != nil {
			return msg, err
		}
		ava = images.ClipImgToCircle(ava)
		img.DrawImage(ava, 10, height)
		// 写昵称+ID
		userInfo := ctx.GetStrangerInfo(user.ID, false)
		str := fmt.Sprintf("%s\n%d", strings.TrimSpace(userInfo.Get("nickname").String()), user.ID)
		realIdW, _ := images.MeasureStringDefault(str, fontSize, 1.3)
		if realIdW > float64(idSize) { // 昵称过长，裁剪
			nn := []rune(strings.TrimSpace(userInfo.Get("nickname").String()))
			nn = nn[:int((float64(idSize)/realIdW)*float64(len(nn)))]
			str = fmt.Sprintf("%s\n%d", string(nn), user.ID)
		}
		err = img.PasteStringDefault(str, fontSize, 1.3, float64(10+avaSize+10), float64(height+20), float64(idSize))
		if err != nil {
			return msg, err
		}
		// 画线
		value := strconv.FormatFloat(user.Favor, 'f', 2, 64)
		length := user.Favor / users[0].Favor
		if key == "wealth" {
			value = strconv.FormatFloat(RealCoin(user.Wealth), 'f', 0, 64) + Unit()
			length = user.Wealth / users[0].Wealth
		}
		lineY := (float64(avaSize) - lineHeight) / 2.0
		img.SetHexColor("#74c0fc")
		img.DrawRoundedRectangle(float64(10+avaSize+10+idSize), float64(height)+lineY, lineLength*length, lineHeight, 5)
		img.Fill()
		err = img.PasteStringDefault(value, fontSize, 1, float64(10+avaSize+10+idSize+10), float64(height)+lineY+10, lineLength)
		if err != nil {
			return msg, err
		}
		height += avaSize + 20
	}
	imgMsg, err := img.GenMessageAuto()
	if err != nil {
		return msg, err
	}
	return imgMsg, nil
}
