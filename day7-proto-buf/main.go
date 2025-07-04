package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
)

var db = map[string]string{ //8001 8002 8003的slow DB都是一样的，可以理解为三个cache server的slow DB数据源是一样的。
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
} 

// 一个Cache Server就是一个Group。每一个cache server都有一个本地的数据源，如果缓存找不到数据了，就去本地数据源里面找。本地数据源在形式上是一个函数，调用这个函数就能完成本地取数据。
func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc( // 这个GetterFunc是对slow DB的访问
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] Search key:", key)

			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 为新创建的Group添加Peer（http Server），并启动http服务。
// addrs是包括自己在内的所有cache的ip+port
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr) // 一般httppool用来handle请求的。
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 在 Go 的 net/http 包中，HTTP 服务器默认会为每个请求启动一个独立的 goroutine（轻量级线程），因此多个请求会并发处理，而不是串行等待前一个请求完成。
func startApiServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { // 一般都是接受请求用指针，响应用值类型。
			key := r.URL.Query().Get("key") // 常见的对请求的前置处理

			v, err := gee.Get(key) // 进来之后都是用8003的gee去获取key的。
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(v.ByteSlice())
		}))

	log.Println("Api Server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

// 运行一次main，开启一个geecache服务器
func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "will launch apiServer?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()

	if api {
		go startApiServer(apiAddr, gee) // 负责接收http请求的，还是ApiServer。注意， apiserver依附于8003
	}

	startCacheServer(addrMap[port], []string(addrs), gee)
}
