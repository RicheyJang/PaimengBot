package help

import (
	"fmt"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils/images"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func formSingleHelpMsg(cmd string, isSuper, isPrimary bool, priority int, blackKeys map[string]struct{}) message.MessageSegment {
	plugins := manager.GetAllPluginConditions()
	// 寻找插件
	var selected *manager.PluginCondition
	for _, plugin := range plugins { // 优先找插件名
		if plugin.Name == cmd && checkPluginCouldShow(plugin, isSuper, isPrimary, priority, blackKeys) {
			selected = plugin
			break
		}
	}
	if selected == nil { // 尝试通过命令
		for _, plugin := range plugins {
			if isCmdContains(plugin, cmd, isSuper) && checkPluginCouldShow(plugin, isSuper, isPrimary, priority, blackKeys) {
				selected = plugin
				break
			}
		}
	}
	if selected == nil {
		return message.Text("没有找到这个功能哦，或在群聊中无法查看功能详情")
	}
	// 插件状态检查
	if _, ok := blackKeys[selected.Key]; ok {
		return message.Text("功能被禁用中")
	}
	// 生成图片 名称|普通用法|超级用户用法
	name := selected.Name
	if selected.AdminLevel > 0 { // 权限等级
		name = fmt.Sprintf("[%d] %s", selected.AdminLevel, name)
	}
	classify := selected.Classify
	if len(classify) > 0 { // 分类
		classify = "（类别：" + classify + "）"
	}
	if isSuper && isPrimary { // Key
		classify = "（插件Key：" + selected.Key + "）"
	}
	usages := name + classify + "\n" + selected.Usage
	if isSuper && len(selected.SuperUsage) > 0 {
		usages += "\n超级用户额外用法：\n" + selected.SuperUsage
	}
	// 计算图片大小并初始化
	fontSize, lineSpace := 20.0, 1.3
	w, h := images.MeasureStringDefault(usages, fontSize, lineSpace)
	w, h = w+20, h+50+10
	img := images.NewImageCtxWithBGRGBA255(int(w), int(h), 255, 255, 255, 255)
	// 名称
	err := img.PasteStringDefault(name+classify, fontSize, lineSpace, 10, 10, w)
	if err != nil {
		log.Warnf("formSingleHelpMsg img err: %v", err)
		return message.Text(usages)
	}
	// 普通用法
	_, nameH := images.MeasureStringDefault(name+classify, fontSize, lineSpace)
	img.PasteLine(10, 10+nameH+10, w-10, 10+nameH+10, 2, "gray")
	err = img.PasteStringDefault(selected.Usage, fontSize, lineSpace, 10, 10+nameH+20, w)
	if err != nil {
		log.Warnf("formSingleHelpMsg img err: %v", err)
		return message.Text(usages)
	}
	// 超级用户用法
	if isSuper && len(selected.SuperUsage) > 0 {
		_, usageH := images.MeasureStringDefault(selected.Usage, fontSize, lineSpace)
		img.PasteLine(10, 10+nameH+20+usageH+10, w-10, 10+nameH+20+usageH+10, 2, "green")
		_ = img.PasteStringDefault("S", fontSize-5, 1, w-20, 10+nameH+20+usageH, fontSize)
		superUsage := "超级用户额外用法：\n" + selected.SuperUsage
		err = img.PasteStringDefault(superUsage, fontSize, lineSpace, 10, 10+nameH+20+usageH+20, w)
		if err != nil {
			log.Warnf("formSingleHelpMsg img err: %v", err)
			return message.Text(usages)
		}
	}
	// 生成回包
	msg, err := img.GenMessageAuto()
	if err != nil {
		log.Warnf("formSingleHelpMsg img err: %v", err)
		return message.Text(usages)
	}
	return msg
}

func isCmdContains(plugin *manager.PluginCondition, cmd string, isSuper bool) bool {
	if isSuper {
		for _, pCmds := range plugin.SuperCmd {
			for _, pCmd := range pCmds {
				if cmd == pCmd {
					return true
				}
			}
		}
	}
	for _, pCmds := range plugin.NormalCmd {
		for _, pCmd := range pCmds {
			if cmd == pCmd {
				return true
			}
		}
	}
	return false
}
