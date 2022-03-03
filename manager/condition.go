package manager

import (
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/robfig/cron/v3"
)

// PluginCondition 插件状况结构，
// Hook类型插件应该只与此结构交互
type PluginCondition struct {
	PluginInfo            // 插件信息（由插件提供，只读）
	Key        string     // 插件Key
	NormalCmd  [][]string // 普通用户专用命令
	SuperCmd   [][]string // 超级用户专用命令

	disabled bool       // 插件是否禁用（全局，针对所有用户），默认false：启用
	schedule *cron.Cron // 定时任务结构
}

// Status 插件全局启用状态
func (c PluginCondition) Status() bool {
	return !c.disabled
}

// Enabled 启用插件
func (c *PluginCondition) Enabled() {
	c.disabled = false
	c.InitialCron()
}

// Disabled 停用插件
func (c *PluginCondition) Disabled() {
	c.disabled = true
	c.StopCron()
}

// InitialCron 初始化定时器
func (c *PluginCondition) InitialCron() {
	if c.schedule == nil {
		cronLogger := utils.NewCronLogger()
		c.schedule = cron.New(cron.WithLogger(cronLogger), // 设置日志
			cron.WithChain(cron.Recover(cronLogger), // panic恢复
				cron.SkipIfStillRunning(cronLogger))) // 若前一任务仍在执行，则跳过当前任务
	}
	c.StartCron()
}

// StartCron 开始所有定时任务
func (c *PluginCondition) StartCron() {
	if c.schedule != nil && c.Status() { // 已有定时任务结构且插件启用
		c.schedule.Start()
	}
}

// StopCron 停止所有定时任务
func (c *PluginCondition) StopCron() {
	if c.schedule != nil {
		c.schedule.Stop()
	}
}
