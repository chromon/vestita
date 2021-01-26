package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 自定义 Hash 函数，默认为 crc32.ChecksumIEEEE
type Hash func(data []byte) uint32

// 一致性哈希算法数据结构，用于操作节点
type Map struct {
	// hash 函数
	hash Hash
	// 虚拟节点倍数
	replicas int
	// 哈希环
	keys []int
	// 虚拟节点与真实节点映射
	// key 虚拟节点 hash 值，value 真实节点的 hash 值
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}

	// hash 函数默认值
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点，根据 replicas 判断是否添加虚拟节点
// keys 为 0 或多个真实节点名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 每一个真实节点对应创建 m.replicas 个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点名，通过编号区分不同虚拟节点
			virNode := strconv.Itoa(i) + key
			hash := int(m.hash([]byte(virNode)))
			// 将计算的新节点哈希添加到环上
			m.keys = append(m.keys, hash)
			// 增加虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	// 将环上的 hash 值排序
	sort.Ints(m.keys)
}

// 由缓存 key 选择真实节点
func (m *Map) Get(key string) string {
	// 没有节点
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// func Search(n int, f func(int) bool) int
	// 采用二分法搜索找到 [0, n) 区间内最小的满足 f(i)==true 的值 i，如果没有该值，函数会返回 n
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 当 idx == len(m.keys) 时，说明所有 keys 都比当前 key hash 小
	// 由于 m.keys 是环状结构，所以应该选择 m.key[0] 节点（使用取余数方式处理）
	// 最后通过 hashMap 映射到真实节点
	return m.hashMap[m.keys[idx % len(m.keys)]]
}
