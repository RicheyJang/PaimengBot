package weather

import (
	"fmt"
	"math"
	"time"

	"github.com/RicheyJang/PaimengBot/utils"

	"github.com/RicheyJang/PaimengBot/utils/client"
)

const freeWeatherURL = "https://v0.yiketianqi.com/api?unescape=1&version=v91&appid=43656176&appsecret=I42og6Lm&ext=&cityid=&city=%v"

// FreeGetWeather 调用免费API获取天气数据
func FreeGetWeather(city string) (Weather, error) {
	// 调用API
	url := fmt.Sprintf(freeWeatherURL, city)
	c := client.NewHttpClient(nil)
	now := time.Now()
	res, err := c.GetGJson(url)
	if err != nil {
		return Weather{}, err
	}
	// 数据解析
	errCode := res.Get("errcode").Int()
	if errCode == 100 {
		return Weather{}, ErrorOfInvalidCity
	}
	if errCode != 0 {
		return Weather{}, fmt.Errorf(res.Get("errmsg").String())
	}
	// 数据格式转换
	update, err := time.Parse("2006-01-02 15:04:05", res.Get("update_time").String())
	if err != nil {
		update = now
	}
	weather := Weather{
		City:      res.Get("city").String(),
		UpdatedAt: update,
	}
	// 各天数据
	for i, data := range res.Get("data").Array() {
		date, err := time.Parse("2006-01-02", data.Get("date").String())
		if err != nil {
			date = now.AddDate(0, 0, i)
		}
		daily := DailyData{
			Date:            date,
			HighTemperature: int(data.Get("tem1").Int()),
			LowTemperature:  int(data.Get("tem2").Int()),
		}
		year, month, day := daily.Date.Date()
		dayD := DetailedData{
			Time:         time.Date(year, month, day, 9, 0, 0, 0, daily.Date.Location()),
			Temperature:  math.MaxInt,
			Weather:      data.Get("wea_day").String(),
			WindClass:    getNumberFromString(data.Get("win_speed").String()),
			WindDirector: data.Get("win.0").String(),
		}
		if utils.StringRealLength(dayD.WindDirector) > 3 {
			if dayD.WindClass > 0 {
				dayD.WindDirector = "风力"
			} else {
				dayD.WindDirector = ""
			}
		}
		nightD := DetailedData{
			Time:         time.Date(year, month, day, 22, 0, 0, 0, daily.Date.Location()),
			Temperature:  math.MaxInt,
			Weather:      data.Get("wea_night").String(),
			WindClass:    0,
			WindDirector: data.Get("win.1").String(),
		}
		if utils.StringRealLength(nightD.WindDirector) > 3 {
			if nightD.WindClass > 0 {
				nightD.WindDirector = "风力"
			} else {
				nightD.WindDirector = ""
			}
		}
		daily.Periods = append(daily.Periods, dayD, nightD)
		weather.Future = append(weather.Future, daily)
	}
	return weather, nil
}

func getNumberFromString(s string) (v int) {
	rs := []rune(s)
	for _, r := range rs {
		if r >= '0' && r <= '9' {
			v *= 10
			v += int(r - '0')
		}
		if r == '-' || r == '级' {
			return
		}
	}
	return
}
