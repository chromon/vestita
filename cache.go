package vestita

import (
	"sync"
	"vestita/lru"
)

// 并发控制
// 封装 lru.Cache 同时添加并发特性
type cache struct {
	// 互斥锁
	mutex sync.Mutex
	lru *lru.Cache
	// 缓存大小
	cacheBytes int64
}

// 并发添加
func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 对象为 nil 时再创建实例 -- 延迟初始化
	// 对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时
	// 主要用于提高性能，并减少程序内存要求
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// 并发读取
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}