package singleflight

import "sync"

/*
缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。
缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起

缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，
造成瞬时DB请求量大、压力骤增

缓存穿透：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，
如果瞬间流量过大，穿透到 DB，导致宕机
*/

// 正在进行中或已经结束的请求
type call struct {
	/* WaitGroup 对象内部有一个计数器，最初从 0 开始，有三个方法：
	 Add(), Done(), Wait() 用来控制计数器的数量。Add(n) 把计数器设置为 n，
	 Done() 每次把计数器 -1 ，wait() 会阻塞代码的运行，直到计数器地值减为 0
	 */
	// 使用 waitgroup 加锁，避免重复请求
	wg sync.WaitGroup
	val interface{}
	err error
}

// 管理不同 key 的请求
type Group struct {
	// 保护 m 不被并发读写
	mutex sync.Mutex
	m map[string]*call
}

// 针对相同的 key，多次调用 Do 方法时，函数 fn 之后调用一次，调用完成后返回
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mutex.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 相关 key 请求以存在
	if c, ok := g.m[key]; ok {
		g.mutex.Unlock()
		// 持续等待，直至结束返回结果
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	// 计数器加 1，加锁
	c.wg.Add(1)
	// 添加 key 到 map 中，表明 key 已经有对应的请求正在等待处理
	g.m[key] = c
	g.mutex.Unlock()

	// 调用 fn 发起请求
	c.val, c.err = fn()
	// 请求结束
	c.wg.Done()

	// 更新 map
	g.mutex.Lock()
	delete(g.m, key)
	g.mutex.Unlock()

	return c.val, c.err
}