package lru

import "container/list"

/*
缓存淘汰策略
	FIFO：先进先出，创建一个队列，新增记录添加到队尾，每次内存不够时，淘汰队首
		但最早添加通常也最常被访问
	LUF：最少使用，需要维护一个按照访问次数排序的队列，每次访问，访问次数加 1，
		队列重新排序，淘汰时选择访问次数最少的即可
		缺点，维护每个记录的访问次数，对内存的消耗是很高的；如果数据的访问模式发生变化，
		LFU 需要较长的时间去适应，某个数据历史上访问次数奇高，但在某个时间点之后几乎不再被访问，
		但因为历史访问次数过高，而迟迟不能被淘汰而占用内存
	LRU：最近最少使用，如果数据最近被访问过，那么将来被访问的概率也会更高。
		维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可
*/

// LRU cache
type Cache struct {
	// 允许使用的最大内存
	// maxBytes 设置为 0，代表不对内存大小设限
	maxBytes int64
	// 当前已使用的内存
	nbytes int64
	// 双向链表
	ll *list.List
	// 缓存，键为 string 值为双向链表对应节点的指针，实际值存放在双向链表中
	cache map[string]*list.Element
	// 移除记录的回调函数
	onEvicted func(key string, value Value)
}

// 双向链表节点的数据类型
type entry struct {
	// 双向链表中仍保存节点 key，是为了在淘汰链表节点时，使用 key 删除 map 中的映射
	key string
	// 值是实现了 Value 接口的任意类型
	value Value
}

// 双向链表节点值类型
type Value interface {
	// 已占用内存大小 bytes
	Len() int
}

// 构造 Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache {
		maxBytes: maxBytes,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 由 key 从字典中找到对应双向链表节点
	if ele, ok := c.cache[key]; ok {
		// 将链表中的节点移动到队尾
		c.ll.MoveToFront(ele)
		// 将链表节点转为 entry
		kv := ele.Value.(*entry)
		// 返回 key 对应的值
		return kv.value, true
	}
	return
}

// 删除，缓存淘汰，移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	// 返回链表最后一个元素（队首）
	ele := c.ll.Back()
	if ele != nil {
		// 删除队首
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 删除 map 中键值对
		delete(c.cache, kv.key)
		// 更新内存占用，减掉 kv 的大小
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

// 添加
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// key 存在，移动节点到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// 已占用内存加上新 value 减掉旧 value 长度
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		// 更新节点值
		kv.value = value
	} else {
		// key 不存在，队尾添加新节点
		ele := c.ll.PushFront(&entry{key, value})
		// map 中添加 key 与新节点映射
		c.cache[key] = ele
		// 更新内存占用
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 判断是否已占用内存超出最大设定，循环移除队首
	// maxBytes 设置为 0，代表不对内存大小设限
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 获取链表长度
func (c *Cache) Len() int {
	return c.ll.Len()
}