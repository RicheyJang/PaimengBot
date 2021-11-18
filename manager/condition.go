package manager

// PluginCondition 插件状况结构，
// Hook类型插件应该只与此结构交互
type PluginCondition struct {
	PluginInfo            // 插件信息（由插件提供，只读）
	NormalCmd  [][]string // 普通用户专用命令
	SuperCmd   [][]string // 超级用户专用命令
}
