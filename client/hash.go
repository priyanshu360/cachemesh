package client

import (
	"hash/fnv"
	"sort"
	"sync"
)

type ConsistentHashRing struct {
	mu         sync.RWMutex
	nodes      map[uint32]string
	vNodeCount int
}

func NewConsistentHashRing(vNodeCount int) *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes:      make(map[uint32]string),
		vNodeCount: vNodeCount,
	}
}

func (r *ConsistentHashRing) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.vNodeCount; i++ {
		vNodeKey := node
		if r.vNodeCount > 1 {
			vNodeKey = node + "-" + string(rune(i))
		}
		hash := fnvHash(vNodeKey)
		r.nodes[hash] = node
	}
}

func (r *ConsistentHashRing) RemoveNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.vNodeCount; i++ {
		vNodeKey := node
		if r.vNodeCount > 1 {
			vNodeKey = node + "-" + string(rune(i))
		}
		hash := fnvHash(vNodeKey)
		delete(r.nodes, hash)
	}
}

func (r *ConsistentHashRing) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.nodes) == 0 {
		return ""
	}

	hash := fnvHash(key)
	keys := make([]uint32, 0, len(r.nodes))
	for k := range r.nodes {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, k := range keys {
		if hash <= k {
			return r.nodes[k]
		}
	}

	return r.nodes[keys[0]]
}

func fnvHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
