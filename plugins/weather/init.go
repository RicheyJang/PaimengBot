package weather

import (
	"fmt"
	"math"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"
	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var info = manager.PluginInfo{
	Name: "天气",
	Usage: `看看今天和未来几天的天气怎么样
用法：
	[城市]天气：查询指定城市未来几天的天气`,
	Classify: "实用工具",
}
var proxy *manager.PluginProxy

func init() {
	proxy = manager.RegisterPlugin(info)
	if proxy == nil {
		return
	}
	proxy.OnRegex(`^(\S{1,10})天气$`, zero.OnlyToMe).SetBlock(true).SetPriority(3).Handle(weatherHandler)
}

type Weather struct {
	City      string
	UpdatedAt time.Time
	Tips      []string    // 数条提示信息，每条为单行不超过15字，仅展示前10条
	Future    []DailyData // 未来的天气，Future[0]为今天
}

type DailyData struct {
	Date            time.Time
	LowTemperature  int // 最低温度
	HighTemperature int // 最高温度
	Periods         []DetailedData
}

type DetailedData struct {
	Time         time.Time
	Temperature  int    // 温度
	Weather      string // 天气状况
	WindClass    int    // 风力级别
	WindDirector string // 风向
}

var ErrorOfInvalidCity = fmt.Errorf("暂不支持该城市或城市不存在")

func weatherHandler(ctx *zero.Ctx) {
	// 当前仅使用免费API
	reg := utils.GetRegexpMatched(ctx)
	if len(reg) <= 1 {
		return
	}
	weather, err := FreeGetWeather(reg[1])
	if err == ErrorOfInvalidCity {
		ctx.Send(ErrorOfInvalidCity.Error())
		return
	} else if err != nil {
		log.Errorf("FreeGetWeather err: %v", err)
		ctx.Send("失败了...")
		return
	}
	// 数据校验
	if len(weather.Future) == 0 || len(weather.Future[0].Periods) == 0 {
		log.Error("Check weather data failed: no valid data")
		ctx.Send("失败了...")
		return
	}
	// 画图
	pic, err := genWeatherPicMsg(weather)
	if err != nil {
		log.Warnf("genWeatherPicMsg err: %v", err)
		data := weather.Future[0]
		str := fmt.Sprintf("%s今日天气：%s\n%d℃~%d℃", weather.City, data.Periods[0].Weather,
			data.LowTemperature, data.HighTemperature)
		if len(data.Periods[0].WindDirector) > 0 {
			str += fmt.Sprintf("\n%s%d级", data.Periods[0].WindDirector, data.Periods[0].WindClass)
		}
		ctx.Send(str)
		return
	}
	ctx.Send(pic)
}

func genWeatherPicMsg(weather Weather) (message.MessageSegment, error) {
	// 规整数据
	if len(weather.Future) > 7 {
		weather.Future = weather.Future[:7]
	}
	// 0. 初始化
	W, H := 20+20+110*len(weather.Future)+50, 520
	if len(weather.Tips) > 0 {
		W += 380
	}
	img := images.NewImageCtxWithBGColor(W, H, "#ebedf2")
	// 1. 大标题
	err := img.PasteStringDefault(weather.City, 34, 1, 60, 10, float64(W))
	if err != nil {
		return message.MessageSegment{}, err
	}
	w, h := images.MeasureStringDefault(weather.City, 34, 1) // w, h 用于各面板定位
	err = img.PasteStringDefault("更新于"+weather.UpdatedAt.Format("15:04"), 20, 1, 60+w+20, 25, float64(W))
	if err != nil {
		return message.MessageSegment{}, err
	}
	h += 30
	// 2. 天气面板
	w = 20
	// 2.1 白色背景
	x, y := w+20, h+20 // x, y 用于各天子面板定位
	img.PasteRoundedRectangle(w, h, float64(20+110*len(weather.Future)), 410, 10, "white")
	// 2.2 今天的浅蓝色背景
	img.PasteRectangle(x, y, 100, 370, "#f8faff")
	x += 50
	// 2.3 各天天气
	_ = img.UseDefaultFont(24)
	img.SetHexColor("#384c78")
	var lows, highs []int
	for _, daily := range weather.Future {
		lows = append(lows, daily.LowTemperature)
		highs = append(highs, daily.HighTemperature)
		img.DrawStringAnchored(getDateString(daily.Date), x, y+5, 0.5, 1)
		if len(daily.Periods) > 0 {
			drawWeatherDetail(img, x, y+40, 110, 65, daily.Periods[0])
		}
		if len(daily.Periods) > 1 {
			drawWeatherDetail(img, x, y+370-65, 110, 65, daily.Periods[len(daily.Periods)-1])
		}
		x += 110
	}
	// 2.4 气温曲线
	drawTemperatureCurve(img, w+20+50, h+20+105, float64(110*(len(weather.Future)-1)), 200, highs, lows)
	w += float64(20+110*len(weather.Future)) + 20
	// 3. 提示面板
	if len(weather.Tips) == 0 { // 无提示
		return img.GenMessageAuto()
	}
	img.PasteRoundedRectangle(w, h, 380, 410, 10, "white")
	y = h + 10
	x = w + 30
	img.PasteCircle(x, y+18, 5, "black")
	err = img.PasteStringDefault("小提示：", 28, 1, x+15, y, 370)
	if err != nil {
		return message.MessageSegment{}, err
	}
	y += 40
	for i, tip := range weather.Tips {
		if i >= 10 {
			break
		}
		img.PasteCircle(x, y+15, 3, "black")
		err = img.PasteStringDefault(tip, 24, 1, x+8, y, 370)
		if err != nil {
			return message.MessageSegment{}, err
		}
		y += 35
	}
	return img.GenMessageAuto()
}

// 绘制气温曲线，左上角定位
func drawTemperatureCurve(img *images.ImageCtx, x, y, w, h float64, series ...[]int) {
	max, min, num := math.MinInt, math.MaxInt, 0
	for _, s := range series {
		for _, v := range s {
			if v > max {
				max = v
			}
			if v < min {
				min = v
			}
		}
		if len(s) > num {
			num = len(s)
		}
	}
	h -= 60
	y += 30
	xInterval, yInterval := w, float64(0)
	if num > 1 {
		xInterval = w / float64(num-1)
	}
	if max-min > 0 {
		yInterval = h / float64(max-min)
	}
	img.Push()
	defer img.Pop()
	img.SetLineWidth(2)
	_ = img.UseDefaultFont(18)
	for i, s := range series {
		if i%2 == 0 {
			img.SetHexColor("#fcc370")
		} else {
			img.SetHexColor("#94ccf9")
		}
		lastX, lastY := 0.0, 0.0
		for j, v := range s {
			nowX, nowY := x+float64(j)*xInterval, y+float64(max-v)*yInterval
			img.DrawCircle(nowX, nowY, 4)
			img.Fill()
			if j != 0 {
				img.DrawLine(lastX, lastY, nowX, nowY)
				img.Stroke()
			}
			sign := 1
			if i%2 == 0 {
				sign = -1
			}
			img.DrawStringAnchored(fmt.Sprintf("%d℃", v), nowX, nowY+float64(sign*10), 0.5, float64(i%2))
			lastX, lastY = nowX, nowY
		}
	}
}

// 绘制天气+风向，x居中y上定位，至少需占用80 * 60
func drawWeatherDetail(img *images.ImageCtx, x, y, W, H float64, detail DetailedData) {
	img.Push()
	defer img.Pop()
	_ = img.UseDefaultFont(26)
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawStringWrapped(detail.Weather, x, y, 0.5, 0, W, 1, gg.AlignCenter)
	if len(detail.WindDirector) > 0 {
		_ = img.UseDefaultFont(16)
		img.SetHexColor("#8a9baf")
		img.SetLineWidth(1)
		img.DrawRoundedRectangle(x-40, y+H-26, 80, 20, 3)
		img.Stroke()
		windStr := fmt.Sprintf("%s", detail.WindDirector)
		if detail.WindClass > 0 {
			windStr += fmt.Sprintf("%d级", detail.WindClass)
		}
		img.DrawStringWrapped(windStr, x, y+H-10,
			0.5, 1, W, 1, gg.AlignCenter)
	}
}

func getDateString(date time.Time) string {
	if date.YearDay() == time.Now().YearDay() {
		return "今天"
	}
	return fmt.Sprintf("%d月%d日", date.Month(), date.Day())
}
