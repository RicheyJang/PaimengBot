package bilibili

import "time"

const (
	SubTypeUp      = "up"
	SubTypeBangumi = "bangumi"
	SubTypeLive    = "live"
)

type Subscription struct {
	ID       int
	SubType  string
	SubUsers string
	// 状态相关：
	BID              int64     // UP主ID、番剧ID、直播间ID
	BangumiLastIndex string    // 番剧：最后一集Index
	DynamicLastTime  time.Time // UP动态：最后一条动态时间戳
	LiveStatus       bool      // 直播间：开播状态
}
