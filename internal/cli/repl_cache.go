package cli

import (
	"sync"
	"time"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	Data     interface{}
	CachedAt time.Time
}

var (
	replCache   = make(map[string]*CacheEntry)
	replCacheMu sync.RWMutex
)

const cacheTTL = 5 * time.Minute

// cacheGet 从缓存读取，过期或不存在返回 nil
func cacheGet(key string) interface{} {
	replCacheMu.RLock()
	defer replCacheMu.RUnlock()
	entry, ok := replCache[key]
	if !ok || time.Since(entry.CachedAt) > cacheTTL {
		return nil
	}
	return entry.Data
}

// cacheSet 写入缓存
func cacheSet(key string, data interface{}) {
	replCacheMu.Lock()
	defer replCacheMu.Unlock()
	replCache[key] = &CacheEntry{Data: data, CachedAt: time.Now()}
}

// InvalidateCache 在写操作后清除指定缓存键
func InvalidateCache(keys ...string) {
	replCacheMu.Lock()
	defer replCacheMu.Unlock()
	for _, k := range keys {
		delete(replCache, k)
	}
}
