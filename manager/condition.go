package manager

import "github.com/robfig/cron/v3"

// PluginCondition 插件状况结构，
// Hook类型插件应该只与此结构交互
type PluginCondition struct {
	PluginInfo            // 插件信息（由插件提供，只读）
	NormalCmd  [][]string // 普通用户专用命令
	SuperCmd   [][]string // 超级用户专用命令

	disabled bool       // 插件是否启用，默认false：启用
	schedule *cron.Cron // 定时任务结构
}

// Enabled 启用插件
func (c *PluginCondition) Enabled() {
	c.disabled = false
	c.StartCron()
}

// Disabled 停用插件
func (c *PluginCondition) Disabled() {
	c.disabled = true
	c.StopCron()
}

// StartCron 开始所有定时任务
func (c *PluginCondition) StartCron() {
	if c.schedule != nil && !c.disabled { // 已有定时任务结构且插件启用
		c.schedule.Start()
	}
}

// StopCron 停止所有定时任务
func (c *PluginCondition) StopCron() {
	if c.schedule != nil {
		c.schedule.Stop()
	}
}
