package test

import (
	"reflect"
	"testing"
	"vestita"
)

func TestGetter(t *testing.T) {
	// 借助 GetterFunc 的类型转换，将匿名回调函数转换成了接口 f Getter
	var f vestita.Getter = vestita.GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}