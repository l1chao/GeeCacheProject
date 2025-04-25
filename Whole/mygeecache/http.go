package mygeecache

import (
	"mygeecache/consistenthash"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self     string // "http://localhost:8001" 用来记录自己的IP和端口
	basePath string // "/_geecache/" 项目名。
	mu       sync.Mutex
	peers    *consistenthash.Map
	httpGetters
}

type HTTPGetter struct {
	baseURL string // e.g. "http://localhost:8001/_geecache/"
}
