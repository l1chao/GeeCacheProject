package mygeecache

import (
	"mygeecache/lru"
	"sync"
)

// 在lru基础上面实现多线程安全的cache。
// 这里的线程安全是最基础的线程安全，仅仅是保证所有的线程同时只能单一访问get和add方法。
// 多线程安全的cache只有add和get，也就是说，会调用add或get直到缓存满。缓存满之后的删除是lru里面已经实现的机制。

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}
