package help

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type blockInfo struct {
	items  []blockItem
	height float64
}

type blockItem struct {
	name  string
	color string
}

type helpSummaryMap map[string]*blockInfo

func formSummaryHelpMsg(isSuper bool, priority int) message.MessageSegment {
	plugins := manager.GetAllPluginConditions()
	defaultClassify := "一般功能"
	hiddenClassify := "被动"
	// 获取所有插件信息
	var helps helpSummaryMap = make(map[string]*blockInfo)
	for _, plugin := range plugins {
		// 过滤
		if plugin.IsSuperOnly && !isSuper {
			continue // 非超级用户
		}
		if plugin.AdminLevel != 0 && (priority == 0 || plugin.AdminLevel < priority) {
			continue // 权限不足
		}
		// 生成项目
		var item blockItem
		item.name = plugin.Name
		item.color = "black"
		if plugin.IsHidden && len(plugin.Classify) != 0 && plugin.Classify != hiddenClassify {
			item.name += "（被动）" // 隐藏且已有其它分类
		}
		if plugin.AdminLevel != 0 {
			item.name = fmt.Sprintf("[%d]", plugin.AdminLevel) + item.name
		}
		if (len(plugin.SuperCmd) > 0 || plugin.IsSuperOnly) && isSuper {
			item.color = "green" // 含超级用户指令
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
			helps[classify] = &blockInfo{items: []blockItem{item}}
		}
	}
	// TODO 生成图片
	//for _,block := range helps {
	//
	//}
	return message.Text(helps)
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
