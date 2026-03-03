package main

import (
	"fmt"
	"time"

	"github.com/priyanshu360/cachemesh/cache"
	"github.com/priyanshu360/cachemesh/config"
	"github.com/priyanshu360/cachemesh/hash"
	"github.com/priyanshu360/cachemesh/node"
	"github.com/priyanshu360/cachemesh/storage"
)

func main() {
	lruStorage := storage.NewLRU(3)
	lru := cache.New(lruStorage, lruStorage, 3)

	lru.Set("a", 1, time.Minute)
	lru.Set("b", 2, time.Minute)
	lru.Set("c", 3, time.Minute)

	val, _ := lru.Get("a")
	fmt.Printf("LRU Get(a): %v\n", val)

	lru.Set("d", 4, time.Minute)
	val, _ = lru.Get("b")
	fmt.Printf("LRU Get(b) after evict: %v\n", val)

	fmt.Printf("LRU Stat: %+v\n", lru.Stat())

	lfuStorage := storage.NewLFU(3)
	lfu := cache.New(lfuStorage, lfuStorage, 3)
	lfu.Set("a", 1, time.Minute)
	lfu.Set("b", 2, time.Minute)
	lfu.Set("c", 3, time.Minute)

	lfu.Get("a")
	lfu.Get("a")
	lfu.Get("b")

	lfu.Set("d", 4, time.Minute)
	val, _ = lfu.Get("c")
	fmt.Printf("LFU Get(c) after evict: %v\n", val)
	fmt.Printf("LFU Stat: %+v\n", lfu.Stat())

	ring := hash.New(100)
	ring.AddNode(hash.NodeInfo{ID: "node1", Addr: "localhost", Port: 8080})
	ring.AddNode(hash.NodeInfo{ID: "node2", Addr: "localhost", Port: 8081})
	ring.AddNode(hash.NodeInfo{ID: "node3", Addr: "localhost", Port: 8082})

	keys := []string{"user:1", "user:2", "user:3", "product:100", "product:200"}
	for _, k := range keys {
		n := ring.GetNode(k)
		fmt.Printf("Key %s -> Node %s\n", k, n)
	}

	cfg := config.Config{
		NodeID:     "node1",
		Addr:       "localhost",
		Port:       8080,
		CacheSize:  1000,
		VNodeCount: 100,
		Eviction:   "lru",
	}

	lruStorage = storage.NewLRU(cfg.CacheSize)
	cacheNode := node.New(cfg, lruStorage, lruStorage)
	_ = cacheNode

	fmt.Printf("\nConsistentHash('key1', 3): %d\n", hash.SimpleHash("key1", 3))
	fmt.Printf("ConsistentHash('key2', 3): %d\n", hash.SimpleHash("key2", 3))

	distCache := node.NewDistributedCache([]hash.NodeInfo{
		{ID: "node1", Addr: "localhost", Port: 8080},
		{ID: "node2", Addr: "localhost", Port: 8081},
	})
	_ = distCache
}
