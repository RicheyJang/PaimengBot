package help

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/fogleman/gg"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func formSummaryHelpMsg(isSuper, isPrimary bool, priority int, blackKeys map[string]struct{}) message.MessageSegment {
	plugins := manager.GetAllPluginConditions()
	defaultClassify := "一般功能"
	hiddenClassify := "被动"
	// 获取所有插件信息
	var helps helpSummaryMap = make(map[string]*blockInfo)
	for _, plugin := range plugins {
		// 过滤
		if !checkPluginCouldShow(plugin, isSuper, isPrimary, priority) {
			continue
		}
		// 生成项目
		var item blockItem
		item.name = plugin.Name
		item.color = "black"
		if _, ok := blackKeys[plugin.Key]; ok {
			item.disabled = true // 插件对该用户或群被禁用
		}
		if plugin.IsHidden && len(plugin.Classify) != 0 && plugin.Classify != hiddenClassify {
			item.name += "（被动）" // 隐藏且已有其它分类
		}
		if plugin.AdminLevel != 0 { // 具有权限要求
			item.name = fmt.Sprintf("[%d] ", plugin.AdminLevel) + item.name
		}
		if (len(plugin.SuperCmd) > 0 || plugin.IsSuperOnly) && (isSuper && isPrimary) {
			item.color = "green" // 含超级用户指令且私聊
		}
		// 分类
		classify := plugin.Classify
		if len(classify) == 0 { // 无分类
			classify = defaultClassify
		}
		if plugin.IsHidden && len(plugin.Classify) == 0 {
			classify = hiddenClassify // 默认隐藏插件的分类为"被动"
		}
		if block, ok := helps[classify]; ok {
			block.items = append(block.items, item)
		} else {
			helps[classify] = &blockInfo{classify: classify, items: []blockItem{item}}
		}
	}
	headTips := "所有功能列表  （划红线的为被禁用功能）\n若想查看某一项功能的详细内容, 请输入：帮助 功能名\n"
	if isSuper && isPrimary {
		headTips += "绿字标识的代表包含超级用户专属内容\n"
	}
	// 生成子图片
	w, h := images.MeasureStringDefault(headTips, 24, 1.3)
	nowH := 10 + h + 20
	i := 0
	for _, block := range helps {
		i += 1
		block.fill(i)
		w = math.Max(block.w, w)
		h += block.h + 40
	}
	w, h = w+30, h+40
	// 提示文字
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	err := img.PasteStringDefault(headTips, 24, 1.3, 10, 10, w)
	if err != nil {
		return message.Text(helps)
	}
	// 一般功能
	if block, ok := helps[defaultClassify]; ok {
		img.DrawImage(block.img, 15, int(nowH))
		nowH += block.h + 40
	}
	// 其它功能
	for c, block := range helps {
		if c == defaultClassify || c == hiddenClassify {
			continue
		}
		img.DrawImage(block.img, 15, int(nowH))
		nowH += block.h + 40
	}
	// 被动功能
	if block, ok := helps[hiddenClassify]; ok {
		img.DrawImage(block.img, 15, int(nowH))
	}
	msg, err := img.GenMessageAuto()
	if err != nil {
		return message.Text(helps)
	}
	return msg
}

type blockInfo struct {
	classify string
	items    []blockItem

	img  image.Image
	w, h float64
}

type blockItem struct {
	name     string
	color    string
	disabled bool
}

type helpSummaryMap map[string]*blockInfo

var bgColors = []string{
	"153,233,242",
	"178,242,187",
	"255,236,153",
	"252,194,215",
	"186,200,255",
}

func (block *blockInfo) fill(num int) {
	fontSize, lineSpace := 24.0, 1.3
	// 计算长宽
	block.w, block.h = images.MeasureStringDefault(block.classify, fontSize, lineSpace)
	for _, item := range block.items {
		w, h := images.MeasureStringDefault(item.name, fontSize, lineSpace)
		block.w = math.Max(block.w, w)
		block.h += 10 + h
	}
	block.w += 20
	block.h += 20
	// 生成图片
	a := 140
	img := images.NewImageCtx(int(block.w), int(block.h))
	defer func() {
		block.img = img.Image()
	}()
	// 背景
	bg := "rgba(" + bgColors[num%len(bgColors)] + "," + strconv.Itoa(a) + ")"
	img.SetColorAuto(bg)
	img.Clear()
	// 文字
	if err := img.UseDefaultFont(fontSize); err != nil {
		return
	}
	nowH := 10.0
	if len(block.items) > 0 { // 类别作为标题
		block.classify += "："
		img.SetColorAuto("black")
		img.DrawStringWrapped(block.classify, 10, nowH, 0, 0, block.w, lineSpace, gg.AlignLeft)
		_, tmpH := img.MeasureString(block.classify)
		nowH += 10 + tmpH
	}
	for _, item := range block.items { // 各个功能（插件）名
		img.SetColorAuto(item.color)
		img.DrawStringWrapped(item.name, 10, nowH, 0, 0, block.w, lineSpace, gg.AlignLeft)
		tmpW, tmpH := img.MeasureString(item.name)
		if item.disabled { // 已禁用插件
			img.PasteLine(10, nowH+tmpH/2+3, 10+tmpW, nowH+tmpH/2+3, 6, "red")
		}
		nowH += 10 + tmpH
	}
}

func (sm helpSummaryMap) String() string {
	res := "全部功能：\n"
	for c, b := range sm {
		res += c + "：\n"
		for _, item := range b.items {
			res += item.name + "\n"
		}
	}
	return res
}
