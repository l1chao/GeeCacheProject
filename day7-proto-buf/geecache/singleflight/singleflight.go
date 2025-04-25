package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{} // 保存请求结果
	err error       // 保存请求结果
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call // 共享数据结构，需要锁起来。
}

// Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
// Do 的fn包裹了一次httpGet请求。
// 现在考虑：多线程同时通过同一g实例，调用Do方法
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {

	g.mu.Lock() // 某一个Do使用了g.mu.Lock()之后，下一个要使用Do，也要调用g.mu.Lock()，由于第一个Do已经调用了Lock()，所以第二个Lock()的调用将会被阻塞。

	if g.m == nil { // 懒汉式调用，第一个进来就调用这个
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok { // 如果有key了
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call) // 所有的重复请求共享同一个call实例
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock() //保护的是 g.m[key]共享数据-------------------------------

	c.val, c.err = fn() //在当前的 Group.Do 实现中，如果第一个请求执行完 fn() 后没有新的并发请求到来，下一次针对相同 key 的请求会重新调用 fn() 方法。
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
