package vestita

// 缓存值封装
// 只读，用来表示缓存值
type ByteView struct {
	// 存储真实的缓存值
	// byte 类型能够支持任意的数据类型的存储
	b []byte
}

// 实现 Value 接口的 Len() 方法
func (v ByteView) Len() int {
	return len(v.b)
}

// 获取返回的副本，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 创建全新对象，切断指针指向，防止缓存值被外部程序修改
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// 以字符串形式返回数据（缓存值）
func (v ByteView) String() string {
	return string(v.b)
}