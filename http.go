package vestita

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 搭建 HTTP Server 提供被其他节点访问功能（基于 HTTP 的节点间通信）

// 节点间通信地址前缀
const defaultBasePath = "/_cache/"

// 节点间 HTTP 通信的核心数据结构
type HTTPPool struct {
	// 当前节点地址，如 http://example.com:8080
	self string
	// 当前节点间通信地址前缀（外部能访问到的地址）
	// http://example.com:8080/_cache/ 开头的请求，就用于节点间的访问
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Server 日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s]", p.self, fmt.Sprintf(format, v...))
}

// 处理 HTTP 请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

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