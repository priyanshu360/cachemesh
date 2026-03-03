# cachemesh

A distributed in-memory cache system built from scratch with consistent hashing for horizontal scaling.

## Features

- **LRU & LFU Eviction** - Pluggable eviction policies
- **Consistent Hashing** - Virtual nodes for load distribution
- **TTL Support** - Per-key expiration
- **Horizontal Scaling** - Add/remove nodes without reshuffling all keys
- **TCP Communication** - Simple protocol for inter-node communication
- **Cross-node Invalidation** - Invalidate keys across all nodes
- **JSON Serialization** - Automatic serialization of any Go type
- **Go Client Library** - Easy integration with your applications

## Installation

```bash
go get github.com/priyanshu360/cachemesh
```

## Quick Start

### Run Server

```bash
# Default config
make run

# With custom port
make run-port PORT=8081

# Or use config file
./cachemesh -config=config.yaml -port=8080
```

### Use Client Library

```go
import "github.com/priyanshu360/cachemesh/client"

func main() {
    // Single node
    c := client.New("localhost:8080")
    defer c.Close()

    ctx := context.Background()

    // Set any value - auto serialized to JSON
    c.Set(ctx, "user:1", User{Name: "John", Email: "john@example.com"}, time.Hour)

    // Get with type safety
    var user User
    c.GetTo(ctx, "user:1", &user)

    // Or get as generic interface{}
    val, _ := c.Get(ctx, "user:1")

    // Check existence
    exists, _ := c.Exist(ctx, "user:1")

    // Delete
    c.Delete(ctx, "user:1")

    // Ping
    c.Ping(ctx)
}
```

### Cluster Mode

```go
// Connect to multiple nodes
cluster := client.NewCluster([]string{
    "localhost:8080",
    "localhost:8081",
    "localhost:8082",
})
defer cluster.Close()

// Keys are automatically routed via consistent hashing
cluster.Set(ctx, "user:1", data, time.Hour)

// Invalidate across all nodes
cluster.Invalidate(ctx, "user:1")
```

## Configuration

Create a `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

cache:
  type: "lru"      # lru or lfu
  size: 1000       # max entries
  evictOn: true    # enable eviction

hash:
  vNodeCount: 100  # virtual nodes for consistent hashing

log:
  level: "info"    # debug, info, warn, error
```

Or use CLI flags:

```bash
./cachemesh -port=9090 -cache-size=5000 -cache=lfu
```

## Docker

```bash
# Build
make docker-build

# Run
make docker-run
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

## Project Structure

```
cachemesh/
├── cmd/server/       # Server entry point
├── client/           # Go client library
│   ├── client.go    # Client & ClusterClient
│   └── hash.go     # Consistent hash ring
├── storage/         # LRU, LFU implementations
├── cache/           # Cache wrapper
├── hash/            # Server-side hash ring
├── node/            # Server implementation
├── config/          # Config types
└── config.yaml      # Default config
```

## Makefile Commands

```bash
make build        # Build the binary
make run          # Run the server
make test         # Run tests
make docker-build # Build Docker image
make docker-run   # Run Docker container
```

## License

MIT
