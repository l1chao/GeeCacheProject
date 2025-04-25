package geecache

import (
	"fmt"
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter // 会为每一个cache server 配置一个getter用来查询指定的slowDB（这是用于缓存中查不到数据的时候指明应该从哪里获取数据）
	mainCache cache
	peers     PeerPicker          // HTTPPool实现了PeerPicker接口。实际使用中，先创建Group，再创建peers，随后才开启HTTPServer。
	loader    *singleflight.Group // 只有一个实例，所有共享这一个实例。
}

// A Getter loads data for a key.
// 这里实现了一个函数实现一个接口，这意味着一个接口变量可以接收一个函数。简单来说，函数不能实现接口，但是可以通过类型重命名来实现。
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) //一个group对应一个cacheServer，所有的cache server都能在这里找到
)

// NewGroup create a new instance of Group
// 创建一个新的group（cacheserver）
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil { // 必须配置一个getter用来告诉数据如果没有的时候应该向哪里要
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock() // 下面的 groups[name] 是共享数据 要上锁

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// Get value for a key from cache
// 调用一个group的Get，就是获取这个group里面的键key对应的值v。
// 获取v有多种情况：1.可以在本group里面直接找到key，那么直接返回即可。2.本地没有key，则去远程的peers（其他group）去找key。3.通过提供的getter函数去找key。难易度是从高到低提升的。
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key) // 如果本地找不到，就调用load去远程调用
}

// 1.去远程的peers的cache找key 2.去远程的slow DB找key。
func (g *Group) load(key string) (value ByteView, err error) {

	// load 完全有可能同时被多个请求同时调用。如果同时调用，就可能引起“缓存击穿”的问题。
	// 下面的Do函数是为了解决“缓存击穿”问题。
	viewi, err := g.loader.Do(key, func() (interface{}, error) { // g.loader只有一个，大家都共享这一个实例
		if g.peers != nil { // g.peers里面有全部的cache server ip+port
			if peer, ok := g.peers.PickPeer(key); ok { // 根据key找到下一个cache server
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				} // 再看key是否在这个server上面。“如果有，则一定在这个server上面”
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key) // 所有的peer的cache里面都没有想要的cache，最后只有到slow DB去找了。
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 填充缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 找slow DB -- 将找到的key加入cache中 -- 返回key
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 用创建group伊始时传进来的Getter来找数据。（在slowDB里面找，getter本来就是用来在找不到数据的时候到slowDB里面找数据的）
	if err != nil {
		return ByteView{}, err

	}

	value := ByteView{b: cloneBytes(bytes)}

	g.populateCache(key, value)

	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}

	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
