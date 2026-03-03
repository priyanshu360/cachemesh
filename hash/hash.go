package hash

import (
	"hash/fnv"
	"sort"
	"sync"
)

type NodeInfo struct {
	ID     string
	Addr   string
	Port   int
	Weight int
}

type ConsistentHashRing struct {
	mu         sync.RWMutex
	nodes      map[uint32]string
	nodeInfo   map[string]NodeInfo
	vNodeCount int
}

func New(vNodeCount int) *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes:      make(map[uint32]string),
		nodeInfo:   make(map[string]NodeInfo),
		vNodeCount: vNodeCount,
	}
}

func (r *ConsistentHashRing) AddNode(node NodeInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodeInfo[node.ID] = node

	for i := 0; i < r.vNodeCount; i++ {
		vNodeKey := node.ID
		if r.vNodeCount > 1 {
			vNodeKey = node.ID + "-" + string(rune(i))
		}
		hash := fnvHash(vNodeKey)
		r.nodes[hash] = node.ID
	}
}

func (r *ConsistentHashRing) RemoveNode(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.nodeInfo, nodeID)

	for i := 0; i < r.vNodeCount; i++ {
		vNodeKey := nodeID
		if r.vNodeCount > 1 {
			vNodeKey = nodeID + "-" + string(rune(i))
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

func (r *ConsistentHashRing) GetNodes() map[string]NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]NodeInfo)
	for id, info := range r.nodeInfo {
		result[id] = info
	}
	return result
}

func fnvHash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func SimpleHash(key string, numNodes int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % numNodes
}
