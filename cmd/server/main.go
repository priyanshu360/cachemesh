package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/priyanshu360/cachemesh/config"
	"github.com/priyanshu360/cachemesh/node"
	"github.com/priyanshu360/cachemesh/storage"
)

var (
	configPath = flag.String("config", "config.yaml", "path to config file")
	port       = flag.Int("port", 0, "server port (overrides config)")
	cacheSize  = flag.Int("cache-size", 0, "cache size (overrides config)")
	cacheType  = flag.String("cache", "", "cache type: lru or lfu (overrides config)")
	help       = flag.Bool("help", false, "show help")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	cfg := loadConfig()

	log.Printf("Starting cachemesh node...")
	log.Printf("Server: %s", cfg.Addr())
	log.Printf("Cache: type=%s, size=%d", cfg.Cache.Type, cfg.Cache.Size)

	n := createNode(cfg)

	if err := n.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Node started successfully on %s", cfg.Addr())

	waitForSignal(n)
}

func loadConfig() *config.Config {
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config from %s: %v", *configPath, err)
		}
	}

	if cfg == nil {
		cfg = config.LoadOrDefault()
	}

	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *cacheSize != 0 {
		cfg.Cache.Size = *cacheSize
	}
	if *cacheType != "" {
		cfg.Cache.Type = *cacheType
	}

	return cfg
}

func createNode(cfg *config.Config) *node.Node {
	var storageImpl storage.Storage
	var evictionPolicy storage.EvictionPolicy

	switch cfg.Cache.Type {
	case "lfu":
		storageImpl = storage.NewLFU(cfg.Cache.Size)
		evictionPolicy = storageImpl.(storage.EvictionPolicy)
	default:
		storageImpl = storage.NewLRU(cfg.Cache.Size)
		evictionPolicy = storageImpl.(storage.EvictionPolicy)
	}

	return node.New(cfg, storageImpl, evictionPolicy)
}

func waitForSignal(n *node.Node) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	n.Stop()
	log.Println("Shutdown complete")
}
