package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash     Hash           // hash函数
	replicas int            // 物理节点重复的倍数
	keys     []int          // Sorted. hashed key。
	hashMap  map[int]string // hashed key到raw key的mapping，可以支持hashed key直接反查raw key
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) { // 被加入进去的，都是节点（物理+虚拟）。加进去的是hash不是原key。  // key == "http://localhost:8001"
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash //如果待查询的缓存的hash在映射之后，已经是最后一个hash环上的元素了，这个时候他就应该映射到m.keys[0]，这是下面返回语句构造原因。
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
