package vestita

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"vestita/consistenthash"
)

// 搭建 HTTP Server 提供被其他节点访问功能（基于 HTTP 的节点间通信）

const (
	// 节点间通信地址前缀
	defaultBasePath = "/_cache/"
	// 默认虚拟节点倍数
	defaultReplicas = 50
)

// 节点间 HTTP 通信的核心数据结构
type HTTPPool struct {
	// 当前节点地址，如 http://example.com:8080
	self string
	// 当前节点间通信地址前缀（外部能访问到的地址）
	// http://example.com:8080/_cache/ 开头的请求，就用于节点间的访问
	basePath string
	// 并发同步
	mutex sync.Mutex
	// 一致性哈希 map，用来根据具体的 key 选择节点
	peers *consistenthash.Map
	// 映射远程节点与对应的 httpGetter 客户端，每个远程节点对应一个 httpGetter
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Server 日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 处理 HTTP 请求（HTTP 服务端功能）
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s -", r.Method, r.URL.Path)

	// /<basepath>/<groupname>/<key>
	// 将 /<groupname>/<key> 以 / 分割，返回的切片最多 2 个子字符串，最后一个子字符串包含未进行切割的部分
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// group 名
	groupName := parts[0]
	// 索引 key
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:" + groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// 实例化了一致性哈希算法，并且添加了传入的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 初始化一致性哈希 Map
	p.peers = consistenthash.New(defaultReplicas, nil)
	// 添加初始节点
	p.peers.Add(peers...)
	// 初始化节点与 HTTPGetter 客户端映射
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		// 为每一个节点创建一个 HTTP 客户端 httpGetter
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 实现 PeerPicker 接口
// 根据具体的 key，选择节点，返回节点对应的 HTTP 客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 获取真实节点的 hash 值
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		// 返回节点对应的 HTTP 客户端
		return p.httpGetters[peer], true
	}

	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// HTTP 客户端类，实现 PeerGetter 接口
type httpGetter struct {
	// 表示将要访问的远程节点的地址，例如 http://example.com/_cache/
	baseURL string
}

// 从对应 group 中查找缓存值
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// func QueryEscape(s string) string
	// 对 s 进行转码使之可以安全的用在 URL 查询里
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	// 发送 http get 请求
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 非 200
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)