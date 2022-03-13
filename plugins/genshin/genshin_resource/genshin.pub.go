package genshin_resource

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func getTodayResourceByGenshinPub(file string) (err error) {
	// 周日特判
	if time.Now().Weekday() == time.Sunday {
		str := "今日素材：周日，什么都能打！"
		img := images.NewImageCtxWithBGColor(1000, 100, resourcePicBGColor)
		err = img.PasteStringDefault(str, 32, 1.3, 30, 40, 500)
		if err != nil {
			return err
		}
		return img.SavePNG(file)
	}
	// 正常工作日
	for i := 0; i < 3; i++ { // 最多尝试3次
		err = tryGetGenshinPubResourceShot(file)
		if err == nil { // 直到成功
			break
		}
	}
	return
}

func tryGetGenshinPubResourceShot(file string) error {
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
		genshinResourceScreenshot(`https://genshin.pub/daily`, `.GSContainer_content_box__1sIXz`, &buf),
	); err != nil {
		log.Warnf("chromedp genshinResourceScreenshot err: %v", err)
		return err
	}
	if err := ioutil.WriteFile(file, buf, 0o644); err != nil {
		log.Warnf("genshinResourceScreenshot write file err: %v", err)
		return err
	}
	return nil
}

// elementScreenshot takes a screenshot of a specific element.
func genshinResourceScreenshot(url, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.EmulateViewport(1500, 1500),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.GSIcon_optimized_icon__7M4-Q`, chromedp.ByQueryAll), // 等待各大头照加载完成
		chromedp.WaitVisible(sel),
		chromedp.Sleep(time.Second), // 额外等待1秒
		chromedp.Evaluate("document.getElementsByClassName('MewBanner_root__3GKl2')[0].remove();", nil),
		chromedp.Evaluate("document.getElementsByClassName('GSContainer_gs_container__2FbUz')[0].setAttribute(\"style\", \"height:1050px\");", nil),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}
