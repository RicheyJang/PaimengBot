package images

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/RicheyJang/PaimengBot/utils"

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

// PasteLine 画线
func (img *ImageCtx) PasteLine(x1, y1, x2, y2, lineWidth float64, colorStr string) {
	img.Push()
	defer img.Pop()
	img.SetColorAuto(colorStr)
	img.DrawLine(x1, y1, x2, y2)
	img.SetLineWidth(lineWidth)
	img.Stroke()
}

// PasteRectangle 画矩形
func (img *ImageCtx) PasteRectangle(x, y, w, h float64, colorStr string) {
	img.Push()
	defer img.Pop()
	img.SetColorAuto(colorStr)
	img.DrawRectangle(x, y, w, h)
	img.Fill()
}

// PasteCircle 画圆
func (img *ImageCtx) PasteCircle(x, y, r float64, colorStr string) {
	img.Push()
	defer img.Pop()
	img.SetColorAuto(colorStr)
	img.DrawCircle(x, y, r)
	img.Fill()
}

// PasteRoundedRectangle 画圆角矩形
func (img *ImageCtx) PasteRoundedRectangle(x, y, w, h, r float64, colorStr string) {
	img.Push()
	defer img.Pop()
	img.SetColorAuto(colorStr)
	img.DrawRoundedRectangle(x, y, w, h, r)
	img.Fill()
}

type Point struct {
	X, Y float64
}

// DrawStar 绘制星星 n: 角数; (x, y): 圆心坐标; r: 圆半径
func (img *ImageCtx) DrawStar(n int, x, y, r float64) {
	points := make([]Point, n)
	for i := 0; i < n; i++ {
		a := float64(i)*2*math.Pi/float64(n) - math.Pi/2
		points[i] = Point{x + r*math.Cos(a), y + r*math.Sin(a)}
	}
	for i := 0; i < n+1; i++ {
		index := (i * 2) % n
		p := points[index]
		img.LineTo(p.X, p.Y)
	}
}

// DrawStringWrapped 重载gg的DrawStringWrapped，支持首尾空格和换行
func (img *ImageCtx) DrawStringWrapped(s string, x, y, ax, ay, width, lineSpacing float64, align gg.Align) {
	lines := img.WordWrap(s, width)

	// sync h formula with MeasureMultilineString
	h := float64(len(lines)) * img.FontHeight() * lineSpacing
	h -= (lineSpacing - 1) * img.FontHeight()

	x -= ax * width
	y -= ay * h
	switch align {
	case gg.AlignLeft:
		ax = 0
	case gg.AlignCenter:
		ax = 0.5
		x += width / 2
	case gg.AlignRight:
		ax = 1
		x += width
	}
	ay = 1
	for _, line := range lines {
		img.DrawStringAnchored(line, x, y, ax, ay)
		y += img.FontHeight() * lineSpacing
	}
}

// WordWrap 重载gg的WordWrap，支持首尾空格和换行
func (img *ImageCtx) WordWrap(s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		fields := utils.SplitOnSpace(line)
		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			w, _ := img.MeasureString(x + fields[i])
			if w > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		//if x != "" { // 空行
		result = append(result, x)
		//}
	}
	//for i, line := range result { // 首尾空格
	//	result[i] = strings.TrimSpace(line)
	//}
	return result
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

// SetColorAuto 根据参数自动识别并设置颜色
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
	img.SetHexColor("#ffffff") // 兜底 纯白
}
