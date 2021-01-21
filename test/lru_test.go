package test

import (
	"reflect"
	"testing"
	"vestita/lru"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := lru.New(int64(0), nil)
	lru.Add("k1", String("v1"))
	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatalf("cache hit k1 = v1 failed")
	}

	if _, ok := lru.Get("k2"); ok {
		t.Fatalf("cache miss k2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap := len(k1 + k2 + k3)
	lru := lru.New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("k1"); ok { // || lru.Len() != 2 {
		t.Fatalf("remove oldest k1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value lru.Value) {
		keys = append(keys, key)
	}
	lru := lru.New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}