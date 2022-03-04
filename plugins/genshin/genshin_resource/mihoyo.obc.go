package genshin_resource

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func getTodayEventByMhyObc(file string) (err error) {
	for i := 0; i < 3; i++ { // 最多尝试3次
		err = tryGetMhyObcTodayEventShot(file)
		if err == nil { // 直到成功
			break
		}
	}
	return
}

func tryGetMhyObcTodayEventShot(file string) error {
	// 创建 context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Infof),
		chromedp.WithDebugf(log.Debugf),
		chromedp.WithErrorf(log.Errorf),
	)
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// 截图
	var buf []byte
	if err := chromedp.Run(ctx,
		genshinTodayEventScreenshot(`https://bbs.mihoyo.com/ys/obc/channel/map/193`, `cal-pc__row cal-pc__row--event`, &buf),
	); err != nil {
		log.Warnf("chromedp genshinTodayEventScreenshot err: %v", err)
		return err
	}
	if err := ioutil.WriteFile(file, buf, 0o644); err != nil {
		log.Warnf("genshinTodayEventScreenshot write file err: %v", err)
		return err
	}
	return nil
}

// elementScreenshot takes a screenshot of a specific element.
func genshinTodayEventScreenshot(url, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.EmulateViewport(1500, 1500),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.cal-event__icon`, chromedp.ByQueryAll), // 等待各大头照加载完成
		chromedp.WaitVisible(sel),
		chromedp.Sleep(time.Second), // 额外等待1秒
		// 放大
		chromedp.Evaluate("document.getElementsByClassName('map-catalog')[0].remove();", nil),
		chromedp.Evaluate(`document.getElementsByClassName("cal-pc__row cal-pc__row--event")[0].style.transform = 'scale(1.2,1.2)'`, nil),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}
