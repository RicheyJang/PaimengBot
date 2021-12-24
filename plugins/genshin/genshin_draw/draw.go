package genshin_draw

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/fogleman/gg"

	"github.com/RicheyJang/PaimengBot/utils/images"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 请求处理函数

func drawOneCard(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) < 2 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("格式不对哦，可以看看帮助"))
		return
	}
	if utils.IsMessageGroup(ctx) {
		ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, drawCards(ctx.Event.UserID, 1, subs[1])...)...)
	} else {
		ctx.Send(drawCards(ctx.Event.UserID, 1, subs[1]))
	}
}

func drawTenCard(ctx *zero.Ctx) {
	subs := utils.GetRegexpMatched(ctx)
	if len(subs) < 2 {
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text("格式不对哦，可以看看帮助"))
		return
	}
	if utils.IsMessageGroup(ctx) {
		ctx.SendChain(append(message.Message{message.At(ctx.Event.UserID)}, drawCards(ctx.Event.UserID, 10, subs[1])...)...)
	} else {
		ctx.Send(drawCards(ctx.Event.UserID, 10, subs[1]))
	}
}

func drawCards(userID int64, num int, name string) message.Message {
	if proxy.LockUser(userID) {
		return message.Message{message.Text("有正在进行的抽卡哦，稍等一下嘛")}
	}
	defer proxy.UnlockUser(userID)
	if num > 80 {
		return message.Message{message.Text("抽的太多啦，少来点")}
	}
	// 获取池子
	if len(name) == 0 {
		name = "常驻"
	}
	reg := regexp.MustCompile(`(.*)\d*`)
	subs := reg.FindStringSubmatch(name)
	if len(subs) < 2 {
		return message.Message{message.Text("没有这个祈愿欸")}
	}
	pools := LoadPoolsByPrefix(subs[1])
	var pool *DrawPool
	for _, p := range pools {
		if p.Name == name {
			pool = &p
			break
		}
	}
	if pool == nil {
		return message.Message{message.Text("没有这个祈愿欸")}
	}
	// 读取上次4、5星
	last4Key := fmt.Sprintf("genshin_draw.%v.4", userID)
	last5Key := fmt.Sprintf("genshin_draw.%v.5", userID)
	last4 := getKVNum(last4Key)
	last5 := getKVNum(last5Key)
	items := simulateRepeatedly(pool, num, &last4, &last5)
	if len(items) == 0 {
		return message.Message{}
	}
	// 记录
	_ = putKVNum(last4Key, last4)
	_ = putKVNum(last5Key, last5)
	// 构造图片
	if len(items) == 1 {
		msg, err := items[0].getImage().GenMessageAuto()
		if err != nil {
			log.Warnf("item image GenMessageAuto err: %v", err)
			return message.Message{message.Text(items[0].String())}
		}
		return message.Message{msg}
	}
	lineNum, colNum := (len(items)-1)/5+1, 5
	if colNum > len(items) {
		colNum = len(items)
	}
	singleSize := items[0].getImage().Image().Bounds().Size()
	w, h := singleSize.X*colNum, singleSize.Y*lineNum
	img := images.NewImageCtx(w, h)
	x, y := 0, 0
	for i, item := range items {
		itemImg := item.getImage()
		img.DrawImage(itemImg.Image(), x, y)
		x += singleSize.X
		if x >= w || (i+1)%colNum == 0 {
			x = 0
			y += singleSize.Y
		}
	}
	// 返回消息
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("item image GenMessageAuto err: %v", err)
		var res string
		for _, item := range items {
			res += item.String() + "\n"
		}
		return message.Message{message.Text(res)}
	}
	return message.Message{msg}
}

// 模拟抽卡函数

func simulateRepeatedly(pool *DrawPool, num int, last4 *uint32, last5 *uint32) []innerItem {
	if pool == nil {
		return nil
	}
	items := make([]innerItem, num)
	for i := range items {
		items[i] = simulateOnce(pool, *last4, *last5)
		*last4 = *last4 + 1
		*last5 = *last5 + 1
		if items[i].star == 4 { // 抽到4星，重置
			*last4 = 0
		}
		if items[i].star == 5 { // 抽到5星，重置
			*last5 = 0
		}
	}
	return items
}

func (item innerItem) getImage() (img *images.ImageCtx) {
	if item.img != nil {
		return item.img
	}
	defer func() {
		item.img = img
	}()
	w, h, fontSize, c := 100, 100, 16.0, "white"
	switch item.star {
	case 3:
		c = "#57aac2"
	case 4:
		c = "#b283c5"
	case 5:
		c = "#d5a050"
	}
	img = images.NewImageCtxWithBGColor(w, h+25, c)
	err := img.UseDefaultFont(fontSize)
	if err != nil {
		return img
	}
	// 贴图
	path := utils.PathJoin(item.dir, fmt.Sprintf("%v.png", item.name))
	bg, err := gg.LoadImage(path)
	if err != nil {
		img.SetRGB(0, 0, 0) // 纯黑色
		img.DrawStringWrapped("暂无图片", float64(w)/2, float64(h)/2, 0.5, 0.5, float64(w), 1, gg.AlignCenter)
	} else {
		sx := float64(w) / float64(bg.Bounds().Size().X)
		sy := float64(h) / float64(bg.Bounds().Size().Y)
		img.Push() // 记录原始状态
		img.Scale(sx, sy)
		img.DrawImage(bg, 0, 0) // 贴图
		img.Pop()               // 恢复原始状态
	}
	// 名称
	img.DrawRectangle(0, float64(h), float64(w+1), 25)
	img.SetColorAuto("#e1e5e7")
	img.Fill()
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawStringWrapped(item.name, float64(w)/2, float64(h)+3, 0.5, 0, float64(w), 1, gg.AlignCenter)
	// 星级
	starW, splitW := 17.0, 1.0
	sumW := starW*float64(item.star) + splitW*float64(item.star-1)
	x, y := float64(w)/2-sumW/2+starW/2, float64(h)-5
	img.SetHexColor("#ffcc33")
	for i := 0; i < item.star; i++ {
		img.DrawStar(5, x, y, starW/2)
		img.Fill()
		x += starW + splitW
	}
	return img
}

func (item innerItem) String() string {
	str := fmt.Sprintf("%v\n", item.name)
	for i := 0; i < item.star; i++ {
		str += "★"
	}
	return str
}
