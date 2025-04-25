package mygeecache

import (
	"mygeecache/singleflight"
	"sync"
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	//peers PeerPicker
	loader *singleflight.Group
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (fn GetterFunc) Get(key string) ([]byte, error) { // 这个地方不能用*GetterFunc
	return fn(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)
