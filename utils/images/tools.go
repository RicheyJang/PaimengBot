package images

import (
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"

	"github.com/RicheyJang/PaimengBot/utils"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/fogleman/gg"

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

// ClipImgToCircle 裁切图像成圆形
func ClipImgToCircle(img image.Image) image.Image {
	w := img.Bounds().Size().X
	h := img.Bounds().Size().Y
	// 计算半径与长宽
	radius := math.Max(float64(w), float64(h)) / 2
	w = int(radius * 2)
	h = w

	dc := gg.NewContext(w, h)
	// 画圆形
	dc.DrawCircle(float64(w/2), float64(h/2), radius)
	// 对画布进行裁剪
	dc.Clip()
	// 加载图片
	dc.DrawImageAnchored(img, w/2, h/2, 0.5, 0.5)
	return dc.Image()
}

// GenQQListMsgWithAva 生成带QQ头像的用户或群（以isUser参数区分）列表
func GenQQListMsgWithAva(data map[int64]string, w float64, isUser bool) (msg message.MessageSegment, err error) {
	var avaReader io.Reader
	avaSize, fontSize, height := 100, 24.0, 10
	img := NewImageCtxWithBGRGBA255(int(w)+avaSize+30, len(data)*(avaSize+20)+30, 255, 255, 255, 255)
	for id, str := range data {
		if isUser {
			avaReader, err = utils.GetQQAvatar(id, avaSize)
		} else {
			avaReader, err = utils.GetQQGroupAvatar(id, avaSize)
		}
		if err != nil {
			return msg, err
		}
		ava, _, err := image.Decode(avaReader)
		ava = ClipImgToCircle(ava)
		if err != nil {
			log.Warnf("Decode avatar err: %v", err)
			return msg, err
		}
		img.DrawImage(ava, 10, height)
		err = img.PasteStringDefault(str, fontSize, 1.3, float64(10+avaSize+10), float64(height+25), w)
		if err != nil {
			return msg, err
		}
		height += avaSize + 20
	}
	imgMsg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("生成图片失败, err: %v", err)
		return msg, err
	}
	return imgMsg, nil
}
