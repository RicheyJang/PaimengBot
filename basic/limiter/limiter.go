package limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// PluginLimiter 插件级限流器，可以区分用户地管理插件的CD限流
type PluginLimiter struct {
	Key string // 插件Key，仅做log所用

	cd    time.Duration
	burst int

	limiters sync.Map
	cdMux    sync.RWMutex
}

// NewPluginLimiter 新建PluginLimiter用于单个插件的限流
func NewPluginLimiter(cd time.Duration, burst int) *PluginLimiter {
	res := &PluginLimiter{
		cd:    cd,
		burst: burst,
	}
	res.ResetCD(cd)
	go res.gc()
	return res
}

// GetCD 获取当前CD
func (pl *PluginLimiter) GetCD() time.Duration {
	pl.cdMux.RLock()
	defer pl.cdMux.RUnlock()
	return pl.cd
}

// ResetCD 重置PluginLimiter的CD时间长度
func (pl *PluginLimiter) ResetCD(cd time.Duration) {
	// 重置CD 会在下次gc之后生效
	pl.cdMux.Lock()
	pl.cd = cd
	pl.cdMux.Unlock()
}

// Allow 判断指定用户(key)能否拿到令牌
func (pl *PluginLimiter) Allow(key int64) bool {
	// 获取subLimiter
	l := pl.getSubLimiter(key)
	if l == nil {
		return false
	}
	// 检查rate
	return l.allow()
}

// ---- 内部方法 ----

// 子Limiter，指定了某个特定用户
type subLimiter struct {
	limiter *rate.Limiter
	lastGet time.Time //上一次获取token的时间
	ttl     time.Duration
}

// 回收过期subLimiter
func (pl *PluginLimiter) gc() {
	for {
		// 等待
		pl.cdMux.RLock()
		defaultTTL := time.Minute
		if pl.cd*3 > defaultTTL {
			defaultTTL = pl.cd * 3
		}
		pl.cdMux.RUnlock()
		time.Sleep(defaultTTL)
		// 回收
		pl.limiters.Range(func(key, value interface{}) bool {
			l, ok := value.(*subLimiter)
			if !ok {
				pl.limiters.Delete(key)
				return true
			}
			if l.lastGet.Add(l.ttl).Before(time.Now()) { // 超时，删除
				pl.limiters.Delete(key)
				return true
			}
			return true
		})
	}
}

// 根据key(用户ID)获取subLimiter
func (pl *PluginLimiter) getSubLimiter(key int64) *subLimiter {
	value, ok := pl.limiters.Load(key) // 从Map中获取subLimiter
	if !ok {                           // 不存在或已超时
		pl.cdMux.RLock()
		l := newSubLimiter(pl.cd, pl.burst) // 新建subLimiter
		pl.cdMux.RUnlock()
		pl.limiters.Store(key, l) // 存储subLimiter
		return l
	}
	l, ok := value.(*subLimiter)
	if ok {
		return l
	}
	return nil
}

// 创建新的subLimiter
func newSubLimiter(cd time.Duration, burst int) *subLimiter {
	return &subLimiter{
		limiter: rate.NewLimiter(rate.Every(cd), burst),
		lastGet: time.Now(),
		ttl:     cd * 3, // 3倍CD作为subLimiter的过期间隔
	}
}

// 判断rate
func (l *subLimiter) allow() bool {
	l.lastGet = time.Now()
	return l.limiter.Allow()
}
