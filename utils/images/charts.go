package images

import (
	"bytes"
	"fmt"
	"image/png"

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
