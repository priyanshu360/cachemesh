package cache

import (
	"github.com/priyanshu360/cachemesh/storage"
	"time"
)

type Cache struct {
	storage        storage.Storage
	evictionPolicy storage.EvictionPolicy
	size           int
}

func New(storage storage.Storage, evictionPolicy storage.EvictionPolicy, size int) *Cache {
	return &Cache{
		storage:        storage,
		evictionPolicy: evictionPolicy,
		size:           size,
	}
}

func (c *Cache) Get(key string) (any, error) {
	return c.storage.Get(key)
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	return c.storage.Set(key, value, ttl)
}

func (c *Cache) Delete(key string) bool {
	return c.storage.Delete(key)
}

func (c *Cache) Exist(key string) bool {
	return c.storage.Exist(key)
}

func (c *Cache) Stat() storage.Stat {
	return c.storage.Stat()
}
