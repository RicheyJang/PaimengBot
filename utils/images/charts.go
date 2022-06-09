package images

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/fogleman/gg"
	"github.com/wcharczuk/go-chart/v2"
)

func (img *ImageCtx) FillDonutChartDefault(title string, values []chart.Value) error {
	// 仅有一个Value时，正常绘图会无法绘制出圆形，所以单独绘制
	if len(values) == 1 {
		img.Push()
		defer img.Pop()
		// 背景 白色
		img.SetColorAuto("white")
		img.Clear()
		// 画圆
		img.SetHexColor("#bac8ff")
		img.DrawCircle(float64(img.Width()/2), float64(img.Height()/2), float64(img.Height()/2)-50)
		img.Fill()
		// 写字
		img.SetRGB(0, 0, 0) // 纯黑色
		if err := img.UseDefaultFont(26); err != nil {
			return err
		}
		img.DrawStringWrapped(title, float64(img.Width()/2), 10, 0.5, 0, float64(img.Width()), 1.3, gg.AlignCenter)
		img.DrawStringWrapped(values[0].Label, float64(img.Width()/2), float64(img.Height()/2), 0.5, 0.5, float64(img.Width()), 1.3, gg.AlignCenter)
		return nil
	}
	// 正常绘图
	font := GetDefaultFont()
	if font == nil {
		return fmt.Errorf("default font is nil")
	}
	pie := chart.DonutChart{
		Title:  title,
		Font:   font,
		Width:  img.Width(),
		Height: img.Height(),
		Values: values,
	}
	// 保存环形图
	imgBuff := bytes.NewBuffer(nil) // 结果缓冲区
	err := pie.Render(chart.PNG, imgBuff)
	if err != nil {
		return err
	}
	pieImg, err := png.Decode(imgBuff)
	if err != nil {
		return err
	}
	img.DrawImage(pieImg, 0, 0)
	return nil
}

type UserValue struct {
	ID       int64
	Nickname string
	Value    float64
}

// FillUserRankDefault 生成默认用户排名图，图片宽度请设为730，排行高度 = len(users)*(120)+65
func (img *ImageCtx) FillUserRankDefault(title string, users []UserValue, unit string) error {
	avaSize, idSize, lineLength, lineHeight, fontSize, height, maxValue := 100, 180, 380.0, 50.0, 24.0, 10, 0.0
	// 标题
	err := img.PasteStringDefault(title, fontSize, 1.3, 15, float64(height), float64(img.Width()))
	if err != nil {
		return err
	}
	height += 35
	// 获取最大值
	for _, user := range users {
		if user.Value > maxValue {
			maxValue = user.Value
		}
	}
	// 画图表
	for _, user := range users {
		// 画头像
		avaReader, err := utils.GetQQAvatar(user.ID, avaSize)
		if err != nil {
			return err
		}
		ava, _, err := image.Decode(avaReader)
		_ = avaReader.Close()
		if err != nil {
			return err
		}
		ava = ClipImgToCircle(ava)
		img.DrawImage(ava, 10, height)
		// 写昵称+ID
		str := fmt.Sprintf("%s\n%d", user.Nickname, user.ID)
		realIdW, _ := MeasureStringDefault(str, fontSize, 1.3)
		if realIdW > float64(idSize) { // 昵称过长，裁剪
			nn := []rune(user.Nickname)
			nn = nn[:int((float64(idSize)/realIdW)*float64(len(nn)))]
			str = fmt.Sprintf("%s\n%d", string(nn), user.ID)
		}
		err = img.PasteStringDefault(str, fontSize, 1.3, float64(10+avaSize+10), float64(height+20), float64(idSize))
		if err != nil {
			return err
		}
		// 画线
		length := user.Value / maxValue
		value := strconv.FormatFloat(user.Value, 'f', 2, 64)
		if strings.HasSuffix(value, ".00") {
			value = value[:len(value)-3]
		}
		value += unit
		lineY := (float64(avaSize) - lineHeight) / 2.0
		img.SetHexColor("#74c0fc")
		img.DrawRoundedRectangle(float64(10+avaSize+10+idSize), float64(height)+lineY, lineLength*length, lineHeight, 5)
		img.Fill()
		err = img.PasteStringDefault(value, fontSize, 1, float64(10+avaSize+10+idSize+10), float64(height)+lineY+10, lineLength)
		if err != nil {
			return err
		}
		height += avaSize + 20
	}
	return nil
}
