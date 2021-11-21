package images

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

// AdjustOpacity 将输入图像m的透明度变为原来的倍数的图像返回。若原来为完成全不透明，则percentage = 0.5将变为半透明
func AdjustOpacity(m image.Image, percentage float64) image.Image {
	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewRGBA64(bounds)
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			colorRgb := m.At(i, j)
			r, g, b, a := colorRgb.RGBA()
			opacity := uint16(float64(a) * percentage)
			//颜色模型转换，至关重要！
			v := newRgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: opacity})
			//Alpha = 0: Full transparent
			rr, _g, bb, aa := v.RGBA()
			newRgba.SetRGBA64(i, j, color.RGBA64{R: uint16(rr), G: uint16(_g), B: uint16(bb), A: uint16(aa)})
		}
	}
	return newRgba
}

// ImageCtx 图片上下文
type ImageCtx struct {
	*gg.Context
}

// NewImageCtxWithBGPath 以背景图片路径创建带有背景图片的图片上下文
func NewImageCtxWithBGPath(w, h int, bgPath string, opacity float64) (*ImageCtx, error) {
	bg, err := gg.LoadImage(bgPath)
	if err != nil {
		return nil, err
	}
	return NewImageCtxWithBG(w, h, bg, opacity), nil
}

// NewImageCtxWithBG 创建带有背景图片的图片上下文，通过opacity设置不透明度
func NewImageCtxWithBG(w, h int, bg image.Image, opacity float64) *ImageCtx {
	if opacity > 0 && opacity < 1 {
		bg = AdjustOpacity(bg, opacity)
	}
	res := NewImageCtx(w, h)
	sx := float64(w) / float64(bg.Bounds().Size().X)
	sy := float64(h) / float64(bg.Bounds().Size().Y)
	// 记录原始状态
	res.Push()
	// 设置背景
	res.Scale(sx, sy)
	res.DrawImage(bg, 0, 0)
	// 恢复原始状态
	res.Pop()
	return res
}

// NewImageCtx 创建全透明背景的图片上下文
func NewImageCtx(w, h int) *ImageCtx {
	dc := gg.NewContext(w, h)
	// 记录原始状态
	dc.Push()
	// 全透明
	dc.SetRGBA(1, 1, 1, 0)
	dc.Clear()
	// 恢复原始状态
	dc.Pop()
	return &ImageCtx{
		Context: dc,
	}
}
