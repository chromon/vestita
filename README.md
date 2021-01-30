# Vestita

vestita is a distributed caching and cache-filling library


## Example

* Data source example

```go
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}
```

* Create cache group
```go
func createGroup() *vestita.Group {
	return vestita.NewGroup("score", 2 << 10,
			vestita.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}
``` 

* Start API server

```go
func startAPIServer(apiAddr string, ves *vestita.Group) {
	http.Handle("/api", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := ves.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
```

* Start caching server

```go
func startCacheServer(addr string, addrs []string, ves *vestita.Group) {
	peers := vestita.NewHTTPPool(addr)
	peers.Set(addrs...)
	ves.RegisterPeers(peers)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
```

* main

```go
apiAddr := "http://localhost:9999"
addrMap := map[int]string {
    8001: "http://localhost:8001",
    8002: "http://localhost:8002",
    8003: "http://localhost:8003",
}

var addrs []string
for _, v := range addrMap {
    addrs = append(addrs, v)
}

ves := createGroup()
go startAPIServer(apiAddr, ves)

startCacheServer(addrMap[8001], []string(addrs), ves)
```

* Test

```shell script
$ curl "http://localhost:9999/api?key=Tom"
630
$ curl "http://localhost:9999/api?key=Abc"
Abc not exist
```

 ## LICENSE
 
 Vestita is distributed under the terms of the GPL-3.0 License.