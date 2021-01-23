package vestita

import (
	"fmt"
	"log"
	"sync"
)

// 负责与外部交互，控制缓存存储和获取
// 获取缓存接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// 定义函数类型 GetterFunc 返回值类型与 Get 方法相同
// 所以 GetterFunc 是一个实现了接口的函数类型，简称为接口型函数
// 当 Getter 作为参数，在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数
type GetterFunc func(key string) ([]byte, error)

// 函数类型 GetterFunc 实现 Getter 接口方法 Get
func (f GetterFunc) Get(key string) ([]byte, error) {
	// 在缓存不存在时，应从数据源获取数据添加到缓存中
	// f 方法用于在缓存不存在时，调用这个函数，得到源数据
	return f(key)
}

// 缓存命名空间
type Group struct {
	// 缓存名
	name string
	// 缓存未命中时获取源数据的回调
	getter Getter
	// 并发缓存
	mainCache cache
}

var (
	mutex sync.RWMutex
	groups = make(map[string]*Group)
)

// 创建 Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mutex.Lock()
	defer mutex.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// 获取特定名称的 Group
func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

// 从 cache 中获取特定 key 的值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Cache] hit")
		return v, nil
	}

	// 未命中时，调用加载源数据
	return g.load(key)
}

// 加载尚未缓存的内容
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// 单机加载本地内容
func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用用户自定义的方法加载源数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	// bytes 是切片，切片不会深拷贝（创建全新对象，仅复制指针）
	// bytes 值是用户自定义方法返回的（用户侧可操作修改）
	// 所以用 cloneBytes 创建全新对象，防止缓存值被外部程序修改
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 将源数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}