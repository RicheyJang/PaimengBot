package genshin_draw

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func updateAllForce(ctx *zero.Ctx) {
	err := updateAll()
	if err != nil {
		ctx.Send("更新失败了...")
	} else {
		ctx.Send("更新成功")
	}
}

func updateAll() error {
	err := utils.GoAndWait(updateCharacterPicture, updateWeaponPicture, updatePoolData)
	if err != nil {
		log.Errorf("原神卡池数据更新失败：%v", err)
		return err
	}
	log.Infof("原神卡池数据更新成功")
	return nil
}

const genshinWeaponSrcURL = "https://wiki.biligame.com/ys/%E6%AD%A6%E5%99%A8%E5%9B%BE%E9%89%B4"
const genshinCharacterSrcURL = "https://wiki.biligame.com/ys/%E8%A7%92%E8%89%B2"
const genshinPoolSrcURL = "https://wiki.biligame.com/ys/%E7%A5%88%E6%84%BF"

func updateCharacterPicture() error {
	if _, err := utils.MakeDir(GenshinPoolPicDir); err != nil {
		return err
	}
	// 请求
	c := client.NewHttpClient(nil)
	reader, err := c.GetReader(genshinCharacterSrcURL)
	if err != nil {
		return err
	}
	defer reader.Close()
	// 解析
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return err
	}
	doc.Find(".resp-tabs-container>.resp-tab-content:first-child").Find(".itemhover.home-box-tag").Each(func(i int, s *goquery.Selection) {
		sub := s.Find(".home-box-tag-1 a").First()
		str, _ := sub.Attr("title") // 人物名字
		path := utils.PathJoin(GenshinPoolPicDir, fmt.Sprintf("%v.png", str))
		url, _ := sub.Find("img").Attr("src") // 大头照.png
		err = client.DownloadToFile(path, url, 2)
		if err != nil {
			log.Warnf("原神%v图像下载失败", str)
		}
	})
	return nil
}

func updateWeaponPicture() error {
	if _, err := utils.MakeDir(GenshinPoolPicDir); err != nil {
		return err
	}
	// 请求
	c := client.NewHttpClient(nil)
	reader, err := c.GetReader(genshinWeaponSrcURL)
	if err != nil {
		return err
	}
	defer reader.Close()
	// 解析
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return err
	}
	doc.Find("#CardSelectTr tr").Find("td:first-child a").Each(func(i int, s *goquery.Selection) {
		str, _ := s.Attr("title") // 人物名字
		path := utils.PathJoin(GenshinPoolPicDir, fmt.Sprintf("%v.png", str))
		url, _ := s.Find("img").Attr("src") // 大头照.png
		err = client.DownloadToFile(path, url, 2)
		if err != nil {
			log.Warnf("原神%v图像下载失败", str)
		}
	})
	return nil
}

func updatePoolData() error {
	if _, err := utils.MakeDir(GenshinDrawPoolDir); err != nil {
		return err
	}
	// 请求
	c := client.NewHttpClient(nil)
	reader, err := c.GetReader(genshinPoolSrcURL)
	if err != nil {
		return err
	}
	defer reader.Close()
	// 解析
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	pools := make(map[int][]DrawPool)
	// 常驻
	normalPool := DrawPool{
		Type:         PoolNormal,
		EndTimestamp: time.Now().AddDate(1000, 0, 0).Unix(),
		Title:        "「奔行世间」常驻祈愿",
		PicURL:       "",
	}
	normalTable := doc.Find(".foldContent>table.wikitable").First()
	normalTable.Find("tr:nth-of-type(1)").First().Find("td>div").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Children().First().Attr("title")
		normalPool.Normal5Character = append(normalPool.Normal5Character, title)
	})
	normalTable.Find("tr:nth-of-type(2)").First().Find("td>div").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Children().First().Attr("title")
		normalPool.Normal4 = append(normalPool.Normal4, title)
	})
	normalTable.Find("tr:nth-last-of-type(3)").First().Find("td>div").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Children().First().Attr("title")
		normalPool.Normal5Weapon = append(normalPool.Normal5Weapon, title)
	})
	normalTable.Find("tr:nth-last-of-type(2)").First().Find("td>div").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Children().First().Attr("title")
		normalPool.Normal4 = append(normalPool.Normal4, title)
	})
	normalTable.Find("tr:nth-last-of-type(1)").First().Find("td>div").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Children().First().Attr("title")
		normalPool.Normal3 = append(normalPool.Normal3, title)
	})
	pools[PoolNormal] = []DrawPool{normalPool}
	// up池
	doc.Find(".mw-parser-output>table.wikitable").Last().Find("table.wikitable").Each(func(i int, s *goquery.Selection) {
		pool := parseSinglePoolByTable(s, &normalPool) // 解析单个池子
		if pool.EndTimestamp >= now {                  // 当前有效池子
			log.Infof("get genshin pool: %v, end=%v", pool.Title, time.Unix(pool.EndTimestamp, 0))
			pools[pool.Type] = append(pools[pool.Type], pool)
		}
	})
	for tp, ps := range pools {
		err := SavePools(tp, ps)
		if err != nil {
			log.Errorf("savePools tp=%v failed: %v", tp, err)
		}
	}
	return nil
}

func parseSinglePoolByTable(s *goquery.Selection, normalPool *DrawPool) (pool DrawPool) {
	// 结束时间 -> 第2行
	timeStr := s.Find("tr:nth-of-type(2)").First().Find("td").Text()
	indexWave := strings.IndexRune(timeStr, '~')
	if indexWave < 0 {
		log.Warnf("parseSinglePoolByTable failed: duration no '~'")
		return
	}
	endTime, err := parsePoolTime(timeStr[indexWave+1:])
	if err != nil {
		log.Warnf("parseSinglePoolByTable failed: parsePoolTime failed end=%v,err=%v", strings.TrimSpace(timeStr[indexWave+1:]), err)
		return
	}
	pool.EndTimestamp = endTime.Unix()
	startTime, err := parsePoolTime(timeStr[:indexWave])
	if err == nil && startTime.After(time.Now()) { // 卡池还未开始
		pool.EndTimestamp = 1
	}
	// 标题+类型+URL -> 首行
	sub := s.Find("th.ys-qy-title img")
	pool.Title, _ = sub.Attr("title")
	pool.PicURL, _ = sub.Attr("src")
	if strings.Contains(pool.Title, "角色") {
		pool.Type = PoolCharacter
	} else if strings.Contains(pool.Title, "武器") {
		pool.Type = PoolWeapon
	}
	// UP 5 星 -> 第3行
	s.Find("tr:nth-of-type(3)").First().Find("td a").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Attr("title")
		pool.Limit5 = append(pool.Limit5, title)
	})
	// UP 4 星 -> 第4行
	s.Find("tr:nth-of-type(4)").First().Find("td a").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Attr("title")
		pool.Limit4 = append(pool.Limit4, title)
	})
	// 其它
	pool.Normal3 = normalPool.Normal3
	pool.Normal4 = utils.DeleteStringInSlice(normalPool.Normal4, append(pool.Limit4, proxy.GetConfigStrings("skip.normal4")...)...)
	if pool.Type == PoolCharacter {
		pool.Normal5Character = utils.DeleteStringInSlice(normalPool.Normal5Character, pool.Limit5...)
	} else if pool.Type == PoolWeapon {
		pool.Normal5Weapon = utils.DeleteStringInSlice(normalPool.Normal5Weapon, pool.Limit5...)
	}
	return
}

var poolTimeLayouts = []string{
	"2006/01/02 15:04",
	"2006/01/02 15:04:05",
	"2006/1/2 15:04:05",
	"2006/1/2 15:04",
	"2006/01/02",
	"2006/1/2",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-1-2 15:04:05",
	"2006-1-2 15:04",
	"2006-01-02",
	"2006-1-2",
}

func parsePoolTime(value string) (tm time.Time, err error) {
	for _, layout := range poolTimeLayouts {
		tm, err = time.ParseInLocation(layout, strings.TrimSpace(value), time.Local)
		if err == nil {
			return tm, nil
		}
	}
	if err == nil {
		return tm, fmt.Errorf("unexpected error")
	}
	return
}
