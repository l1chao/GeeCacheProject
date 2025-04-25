package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// 都应该是阻塞在call那个地方。
// MyDo 实际的过程是：
// 等1号请求完成请求，在1号处理过程中，所有其他请求等待。1号完成的下一时刻，所有后续请求得到1号处理结果，返回并结束函数调用。而1号请求则是最后才结束调用的请求。、
// 问题：
// 1.1号请求与其他请求的处理是不平等的。1号需要将自己的请求结果“借给”xdm，1号本身也要返回请求结果。
// 2. 实际上缓解的程度：在fn()函数执行过程中的同样请求的调用，是会被阻塞的。
func (g *Group) Do(key string, fn func() (val interface{}, err error)) (interface{}, error) {
	g.mu.Lock() // 用来准备call的。

	//下面这个不是必须的
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok { // 冲进来是想获取call
		g.mu.Unlock() // 进入这个循环也是一个一个进来的
		c.wg.Wait()   // 被卡在了这里
		return c.val, c.err
	}

	c := new(call)
	g.m[key] = c
	c.wg.Add(1)
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// 先取关键处，然后再来思考。似乎一切都是有一种从简单到复杂的趋势的。
