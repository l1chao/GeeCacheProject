package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
// 一共实现4个操作：查 改增 删
// 缓存都是通过键值对的形式存储的。
type Cache struct { // Cache首字母大写，是为了让cache.go使用
	maxBytes int64                    // 最大存储的byte数目
	nbytes   int64                    // 已存储的byte数目（包括key的byte数目+value的byte数目）
	ll       *list.List               // 元素是entry。不用链表不能完成“访问时更新”。
	cache    map[string]*list.Element // 这是键值对的键到值的映射。有了这个，可以键直接访问值
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value) // 删除某元素之后，可能需要对于被删除元素进行操作。（比如打印被删除元素）
}

// entry是节点数据类型
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache.
// 如果是已经存在的key，排头+更新值
// 否则排头存入新key
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 该项目里面约定：front是队尾，从front入，从tail出。国外翻译直接，所以首进尾出。
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len()) // 对于ASCII字符来说，一个字节就是一个字符。len获得的就是字节数目。
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes { // 这里的maxBytes!=0的检验是便于测试验证，实际上不应该有!=0的验证
		c.RemoveOldest()
	}
}

// Get looks up a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    // 查的时候也要排到前面
		kv := ele.Value.(*entry) // 这里是强制类型转换
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele) // 链表去除

		kv := ele.Value.(*entry)
		delete(c.cache, kv.key) // map去除

		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())

		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
