package cache

import (
	"encoding/json"

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

func (c *Cache) Get(key string) ([]byte, error) {
	return c.storage.Get(key)
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.storage.Set(key, data, ttl)
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

func (c *Cache) GetAs(key string, dest any) error {
	data, err := c.storage.Get(key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, dest)
}
