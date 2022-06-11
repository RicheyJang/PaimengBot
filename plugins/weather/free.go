package weather

import (
	"encoding/xml"
	"fmt"
	"math"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/client"
)

type etouchWeather struct {
	XMLName     xml.Name          `xml:"resp"`
	Error       string            `xml:"error"`
	City        string            `xml:"city"`
	UpdatedTime string            `xml:"updatetime"`
	NextDays    []etouchDailyData `xml:"forecast>weather"`
	Tips        []etouchTip       `xml:"zhishus>zhishu"`
}

type etouchDailyData struct {
	Date            string             `xml:"date"`
	LowTemperature  string             `xml:"low"`
	HighTemperature string             `xml:"high"`
	Day             etouchDetailedData `xml:"day"`
	Night           etouchDetailedData `xml:"night"`
}

type etouchDetailedData struct {
	Weather      string `xml:"type"`      // 天气状况
	WindClass    string `xml:"fengli"`    // 风力级别
	WindDirector string `xml:"fengxiang"` // 风向
}

type etouchTip struct {
	XMLName xml.Name `xml:"zhishu"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
	Detail  string   `xml:"detail"`
}

const freeWeatherURL = "http://wthrcdn.etouch.cn/WeatherApi?city=%v"

// FreeGetWeather 调用免费API获取天气数据
func FreeGetWeather(city string) (Weather, error) {
	// 调用API
	url := fmt.Sprintf(freeWeatherURL, city)
	c := client.NewHttpClient(nil)
	now := time.Now()
	r, err := c.GetReader(url)
	if err != nil {
		return Weather{}, err
	}
	defer r.Close()
	// 数据解析
	d := xml.NewDecoder(r)
	var etouch etouchWeather
	err = d.Decode(&etouch)
	if err != nil {
		return Weather{}, err
	}
	if etouch.Error == "invalid city" {
		return Weather{}, ErrorOfInvalidCity
	}
	if len(etouch.Error) != 0 {
		return Weather{}, fmt.Errorf(etouch.Error)
	}
	// 数据格式转换
	update, err := time.Parse("15:04", etouch.UpdatedTime)
	if err != nil {
		update = now
	}
	weather := Weather{
		City:      etouch.City,
		UpdatedAt: update,
	}
	for _, tip := range etouch.Tips {
		weather.Tips = append(weather.Tips, fmt.Sprintf("%s：%s", tip.Name, tip.Value))
	}
	// 各天数据
	for i, data := range etouch.NextDays {
		daily := DailyData{
			Date:            now.AddDate(0, 0, i),
			LowTemperature:  getNumberFromString(data.LowTemperature),
			HighTemperature: getNumberFromString(data.HighTemperature),
		}
		year, month, day := daily.Date.Date()
		dayD := DetailedData{
			Time:         time.Date(year, month, day, 9, 0, 0, 0, daily.Date.Location()),
			Temperature:  math.MaxInt,
			Weather:      data.Day.Weather,
			WindClass:    getNumberFromString(data.Day.WindClass),
			WindDirector: data.Day.WindDirector,
		}
		nightD := DetailedData{
			Time:         time.Date(year, month, day, 22, 0, 0, 0, daily.Date.Location()),
			Temperature:  math.MaxInt,
			Weather:      data.Night.Weather,
			WindClass:    getNumberFromString(data.Night.WindClass),
			WindDirector: data.Night.WindDirector,
		}
		daily.Periods = append(daily.Periods, dayD, nightD)
		weather.Future = append(weather.Future, daily)
	}
	return weather, nil
}

func getNumberFromString(s string) (v int) {
	rs := []rune(s)
	found := false
	for _, r := range rs {
		if r > '0' && r < '9' {
			v *= 10
			v += int(r - '0')
			found = true
		} else if found {
			break
		}
	}
	return
}
