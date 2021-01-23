package test

import (
	"fmt"
	"log"
	"testing"
	"vestita"
)

var db = map[string]string {
	"a": "10",
	"b": "20",
	"c": "30",
}

func TestGetSrc(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	score := vestita.NewGroup("score", 2<<10,
			vestita.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := score.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := score.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := score.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}