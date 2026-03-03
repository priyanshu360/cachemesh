package config

type Config struct {
	NodeID     string
	Addr       string
	Port       int
	CacheSize  int
	VNodeCount int
	Eviction   string // "lru" or "lfu"
}

type NodeConfig struct {
	ID     string
	Addr   string
	Port   int
	Weight int
}
