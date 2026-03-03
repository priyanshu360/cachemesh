# cachemesh

A distributed in-memory cache system built from scratch with consistent hashing for horizontal scaling.

## Features

- **LRU & LFU Eviction** - Pluggable eviction policies
- **Consistent Hashing** - Virtual nodes for load distribution
- **TTL Support** - Per-key expiration
- **Horizontal Scaling** - Add/remove nodes without reshuffling all keys
- **TCP Communication** - Simple protocol for inter-node communication
- **Cross-node Invalidation** - Invalidate keys across all nodes

## Installation

```bash
go get github.com/priyanshu360/cachemesh
```

## Quick Start

### Local Cache

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/priyanshu360/cachemesh/cache"
    "github.com/priyanshu360/cachemesh/storage"
)

func main() {
    // LRU cache with 1000 entries
    lruStorage := storage.NewLRU(1000)
    c := cache.New(lruStorage, lruStorage, 1000)
    
    c.Set("user:1", map[string]string{"name": "John"}, time.Hour)
    val, _ := c.Get("user:1")
    fmt.Println(val)
}
```

### Distributed Cache

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/priyanshu360/cachemesh/node"
    "github.com/priyanshu360/cachemesh/hash"
)

func main() {
    // Client connects to cluster
    dist := node.NewDistributedCache([]hash.NodeInfo{
        {ID: "node1", Addr: "localhost", Port: 8080},
        {ID: "node2", Addr: "localhost", Port: 8081},
        {ID: "node3", Addr: "localhost", Port: 8082},
    })
    
    dist.Set("user:1", data, time.Hour)
    val, _ := dist.Get("user:1")
}
```

## Architecture

```
                    Client Request
                          │
                          ▼
            ┌─────────────────────────┐
            │  Consistent Hash Ring   │
            │  (key → node mapping)  │
            └─────────────────────────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
         ┌────────┐ ┌────────┐ ┌────────┐
         │ Node 1 │ │ Node 2 │ │ Node 3 │
         │ LRU/   │ │ LRU/   │ │ LRU/   │
         │ LFU    │ │ LFU    │ │ LFU    │
         └────────┘ └────────┘ └────────┘
```

## Packages

| Package | Description |
|---------|-------------|
| `storage` | LRU, LFU implementations |
| `cache` | Main cache wrapper |
| `hash` | Consistent hashing ring |
| `node` | Server & client |
| `config` | Configuration types |

## Run Tests

```bash
go test ./...
```

## License

MIT
