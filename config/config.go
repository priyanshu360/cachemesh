package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Cache  CacheConfig  `yaml:"cache"`
	Hash   HashConfig   `yaml:"hash"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type CacheConfig struct {
	Type    string `yaml:"type"`    // "lru" or "lfu"
	Size    int    `yaml:"size"`    // max entries
	EvictOn bool   `yaml:"evictOn"` // enable eviction
}

type HashConfig struct {
	VNodeCount int `yaml:"vNodeCount"` // virtual nodes
}

type LogConfig struct {
	Level string `yaml:"level"` // "debug", "info", "warn", "error"
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) LogLevel() string {
	if c.Log.Level == "" {
		return "info"
	}
	return c.Log.Level
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadOrDefault() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Cache: CacheConfig{
			Type:    "lru",
			Size:    1000,
			EvictOn: true,
		},
		Hash: HashConfig{
			VNodeCount: 100,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}
