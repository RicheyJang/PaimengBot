package images

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
)

// PasteStringDefault 以默认方式（默认字体、黑色、Wrapped、左上角定位、文字居左）贴文字
func (img *ImageCtx) PasteStringDefault(str string, fontSize, lineSpace float64, x, y, width float64) error {
	img.Push()
	defer img.Pop()
	if err := img.UseDefaultFont(fontSize); err != nil {
		return err
	} // 默认字体
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawStringWrapped(str, x, y, 0, 0, width, lineSpace, gg.AlignLeft)
	return nil
}

func (img *ImageCtx) PasteLine(x1, y1, x2, y2, lineWidth float64, colorStr string) {
	img.Push()
	defer img.Pop()
	img.SetColorAuto(colorStr)
	img.DrawLine(x1, y1, x2, y2)
	img.SetLineWidth(lineWidth)
	img.Stroke()
}

var colorMap map[string]string = map[string]string{
	"white":  "#ffffff",
	"black":  "#000000",
	"gray":   "#a4b0be",
	"red":    "#e74c3c",
	"blue":   "#3498db",
	"green":  "#2ecc71",
	"yellow": "#ffd43b",
}

func (img *ImageCtx) SetColorAuto(colorStr string) {
	if res, ok := colorMap[colorStr]; ok {
		img.SetHexColor(res)
		return
	}
	if strings.HasPrefix(colorStr, "#") {
		img.SetHexColor(colorStr)
		return
	}
	colorStr = strings.ToLower(colorStr)
	if strings.HasPrefix(colorStr, "rgb") {
		colorStr = strings.ReplaceAll(strings.ReplaceAll(colorStr, " ", ""), "\t", "")
		reg := regexp.MustCompile("rgba?\\((\\d{1,3}),(\\d{1,3}),(\\d{1,3})(,\\d{1,3}\\.?\\d*)?\\)")
		sub := reg.FindStringSubmatch(colorStr)
		if len(sub) <= 4 {
			return
		}
		r, _ := strconv.ParseInt(sub[1], 10, 32)
		g, _ := strconv.ParseInt(sub[2], 10, 32)
		b, _ := strconv.ParseInt(sub[3], 10, 32)
		var a int64 = 255
		if len(sub[4]) > 0 {
			sub[4] = sub[4][1:]
		}
		if strings.Contains(sub[4], ".") {
			sub[4] += "0"
			fa, _ := strconv.ParseFloat(sub[4], 32)
			a = int64(255.0 * fa)
		} else if len(sub[4]) > 0 {
			a, _ = strconv.ParseInt(sub[4], 10, 32)
		}
		img.SetRGBA255(int(r), int(g), int(b), int(a))
		return
	}
}
