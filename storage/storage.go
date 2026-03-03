package storage

import (
	"container/list"
	"sync"
	"time"
)

type Stat struct {
	Hit      int
	Miss     int
	MemAvail int
	MemUsed  int
}

type EvictionPolicy interface {
	Access(key string)
	Evict() (key string, ok bool)
	Reset()
}

type Storage interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) (flag bool)
	Exist(key string) (flag bool)
	Stat() Stat
}

type LRUCache struct {
	capacity int
	mu       sync.RWMutex
	data     map[string][]byte
	ttl      map[string]time.Time
	order    *list.List
	items    map[string]*list.Element
	stats    Stat
}

func NewLRU(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		data:     make(map[string][]byte),
		ttl:      make(map[string]time.Time),
		order:    list.New(),
		items:    make(map[string]*list.Element),
		stats:    Stat{MemAvail: capacity},
	}
}

func (c *LRUCache) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		if expiry, hasTTL := c.ttl[key]; hasTTL && time.Now().After(expiry) {
			c.order.Remove(el)
			delete(c.data, key)
			delete(c.ttl, key)
			delete(c.items, key)
			c.stats.Miss++
			return nil, nil
		}
		c.order.MoveToFront(el)
		c.stats.Hit++
		return c.data[key], nil
	}
	c.stats.Miss++
	return nil, nil
}

func (c *LRUCache) Set(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		c.data[key] = value
		if ttl > 0 {
			c.ttl[key] = time.Now().Add(ttl)
		}
		return nil
	}

	for len(c.data) >= c.capacity {
		if key, ok := c.Evict(); ok {
			delete(c.data, key)
			delete(c.ttl, key)
			if el, ok := c.items[key]; ok {
				c.order.Remove(el)
				delete(c.items, key)
			}
		} else {
			break
		}
	}

	c.data[key] = value
	if ttl > 0 {
		c.ttl[key] = time.Now().Add(ttl)
	}
	el := c.order.PushFront(key)
	c.items[key] = el
	c.stats.MemUsed++

	return nil
}

func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		delete(c.data, key)
		delete(c.ttl, key)
		c.order.Remove(el)
		delete(c.items, key)
		c.stats.MemUsed--
		return true
	}
	return false
}

func (c *LRUCache) Exist(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; ok {
		if expiry, hasTTL := c.ttl[key]; hasTTL && time.Now().After(expiry) {
			return false
		}
		return true
	}
	return false
}

func (c *LRUCache) Stat() Stat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *LRUCache) Access(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
	}
}

func (c *LRUCache) Evict() (string, bool) {
	if c.order.Len() == 0 {
		return "", false
	}

	el := c.order.Back()
	key := el.Value.(string)

	return key, true
}

func (c *LRUCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string][]byte)
	c.ttl = make(map[string]time.Time)
	c.order = list.New()
	c.items = make(map[string]*list.Element)
	c.stats = Stat{MemAvail: c.capacity}
}

type LFUCache struct {
	capacity int
	mu       sync.RWMutex
	data     map[string][]byte
	ttl      map[string]time.Time
	freq     map[string]int
	minFreq  int
	stats    Stat
}

func NewLFU(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		data:     make(map[string][]byte),
		ttl:      make(map[string]time.Time),
		freq:     make(map[string]int),
		minFreq:  0,
		stats:    Stat{MemAvail: capacity},
	}
}

func (c *LFUCache) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val, ok := c.data[key]; ok {
		if expiry, hasTTL := c.ttl[key]; hasTTL && time.Now().After(expiry) {
			delete(c.data, key)
			delete(c.ttl, key)
			delete(c.freq, key)
			c.stats.Miss++
			return nil, nil
		}
		c.freq[key]++
		c.stats.Hit++
		return val, nil
	}
	c.stats.Miss++
	return nil, nil
}

func (c *LFUCache) Set(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		c.data[key] = value
		c.freq[key]++
		if ttl > 0 {
			c.ttl[key] = time.Now().Add(ttl)
		}
		return nil
	}

	for len(c.data) >= c.capacity {
		c.evictOne()
	}

	c.data[key] = value
	c.freq[key] = 1
	c.minFreq = 1
	if ttl > 0 {
		c.ttl[key] = time.Now().Add(ttl)
	}
	c.stats.MemUsed++

	return nil
}

func (c *LFUCache) evictOne() {
	var evictKey string
	min := c.minFreq

	for k, f := range c.freq {
		if f < min {
			min = f
			evictKey = k
		}
	}

	if evictKey != "" {
		delete(c.data, evictKey)
		delete(c.ttl, evictKey)
		delete(c.freq, evictKey)
		c.stats.MemUsed--
	}
}

func (c *LFUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		delete(c.data, key)
		delete(c.ttl, key)
		delete(c.freq, key)
		c.stats.MemUsed--
		return true
	}
	return false
}

func (c *LFUCache) Exist(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[key]; ok {
		if expiry, hasTTL := c.ttl[key]; hasTTL && time.Now().After(expiry) {
			return false
		}
		return true
	}
	return false
}

func (c *LFUCache) Stat() Stat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *LFUCache) Access(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.freq[key]++
}

func (c *LFUCache) Evict() (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.data) == 0 {
		return "", false
	}

	min := c.minFreq
	var evictKey string
	for k, f := range c.freq {
		if f == min {
			evictKey = k
			break
		}
	}

	if evictKey != "" {
		delete(c.data, evictKey)
		delete(c.ttl, evictKey)
		delete(c.freq, evictKey)
		c.stats.MemUsed--
	}

	return evictKey, evictKey != ""
}

func (c *LFUCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string][]byte)
	c.ttl = make(map[string]time.Time)
	c.freq = make(map[string]int)
	c.minFreq = 0
	c.stats = Stat{MemAvail: c.capacity}
}
