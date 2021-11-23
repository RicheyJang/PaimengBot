package images

import (
	"image"
	"image/color"
	"io/ioutil"

	"github.com/golang/freetype/truetype"
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

// ParseFont 解析字体文件，生成truetype.Font结构
func ParseFont(path string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// MeasureStringDefault 测量str在默认情况（默认字体、分行）下的长宽
func MeasureStringDefault(str string, fontSize, lineSpace float64) (float64, float64) {
	img := NewImageCtx(1, 1)
	_ = img.UseDefaultFont(fontSize)
	return img.MeasureMultilineString(str, lineSpace)
}
