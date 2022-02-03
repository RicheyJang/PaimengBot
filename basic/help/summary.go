package help

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strconv"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/images"

	"github.com/fogleman/gg"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const defaultClassify = "一般功能"
const passiveClassify = "被动"

func formSummaryHelpMsg(isSuper, isPrimary bool, priority int, blackKeys map[string]struct{}) message.MessageSegment {
	plugins := manager.GetAllPluginConditions()
	// 获取所有插件信息
	var helps helpSummaryMap = make(map[string]*blockInfo)
	for _, plugin := range plugins {
		// 过滤
		if !checkPluginCouldShow(plugin, isSuper, isPrimary, priority, blackKeys) {
			continue
		}
		// 生成项目(一个插件)
		var item blockItem
		item.name = plugin.Name
		item.color = "black"
		if _, ok := blackKeys[plugin.Key]; ok {
			item.disabled = true // 插件对该用户或群被禁用
		}
		if plugin.IsPassive && len(plugin.Classify) != 0 && plugin.Classify != passiveClassify {
			item.name += "（被动）" // 被动且已有其它分类
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
		if plugin.IsPassive && len(plugin.Classify) == 0 {
			classify = passiveClassify // 默认被动插件的分类为"被动"
		}
		if block, ok := helps[classify]; ok {
			block.items = append(block.items, item)
		} else {
			helps[classify] = &blockInfo{classify: classify, items: []blockItem{item}}
		}
	}
	headTips := "所有功能列表  （划红线的为被禁用功能）\n" +
		"若想查看某一项功能的详细用法, 请输入：帮助 功能名\n" +
		"大多数功能在群聊中使用时，请加上\"%[1]v\"前缀，例如：%[1]v帮助、%[1]v关闭复读\n"
	headTips = fmt.Sprintf(headTips, utils.GetBotNickname())
	if isSuper && isPrimary {
		headTips += "\n绿字标识的代表包含超级用户专属内容\n某些插件名前方括号内的数字代表最低使用权限等级，参见：帮助 权限鉴权"
	}
	// 生成子图片
	blocks := sortAllBlocks(helps)
	maxH, currentH, cols := 600.0, 0.0, 1 // 最大列高度, 当前列高度, 列数
	w, h := images.MeasureStringDefault(headTips, 24, 1.3)
	tipH := h // 文字高度
	maxSingleW := 0.0
	for i, block := range blocks {
		block.fill(i + 1)
		maxSingleW = math.Max(maxSingleW, block.w)
		currentH += block.h + 40
		if currentH >= maxH { // 需要开新列
			currentH = block.h + 40
			if currentH >= maxH { // 单个框过高
				maxH = currentH + 10
			}
			cols += 1
		}
		blocks[i].colNum = cols
	}
	w = math.Max(w, maxSingleW*float64(cols)+20*float64(cols-1)) // 多列，取最大框宽度的n倍
	w, h = w+30, h+maxH+40
	// 贴提示文字
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	err := img.PasteStringDefault(headTips, 24, 1.3, 10, 20, w)
	if err != nil {
		return message.Text(helps.String())
	}
	// 贴图
	nowX, nowY := 15, 20+int(tipH)+40
	for i, block := range blocks {
		if i > 0 && block.colNum > blocks[i-1].colNum { // 开新列
			nowX += int(maxSingleW) + 20
			nowY = 20 + int(tipH) + 40
		}
		img.DrawImage(block.img, nowX, nowY)
		nowY += int(block.h) + 40
	}
	msg, err := img.GenMessageAuto()
	if err != nil {
		return message.Text(helps.String())
	}
	return msg
}

// 将功能块排序
func sortAllBlocks(helps helpSummaryMap) []*blockInfo {
	var res []*blockInfo
	for _, block := range helps {
		res = append(res, block)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].classify == defaultClassify || res[j].classify == passiveClassify {
			return true
		}
		if res[i].classify == passiveClassify || res[j].classify == defaultClassify {
			return false
		}
		return res[i].classify < res[j].classify
	})
	return res
}

type blockInfo struct {
	classify string
	items    []blockItem

	img    image.Image
	w, h   float64
	colNum int
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
	// 排序
	sort.Slice(block.items, func(i, j int) bool { return block.items[i].name < block.items[j].name })
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
