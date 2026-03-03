package node

import (
	"encoding/json"
	"fmt"
	"github.com/priyanshu360/cachemesh/cache"
	"github.com/priyanshu360/cachemesh/config"
	"github.com/priyanshu360/cachemesh/hash"
	"github.com/priyanshu360/cachemesh/storage"
	"net"
	"time"
)

type Request struct {
	Type  string          `json:"type"` // get, set, delete, exist, stat
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
	TTL   int64           `json:"ttl"` // milliseconds
}

type Response struct {
	Value json.RawMessage `json:"value"`
	Flag  bool            `json:"flag"`
	Stat  storage.Stat
	Error string `json:"error,omitempty"`
}

type Node struct {
	config   *config.Config
	cache    *cache.Cache
	listener net.Listener
}

func New(cfg *config.Config, storage storage.Storage, evictionPolicy storage.EvictionPolicy) *Node {
	c := cache.New(storage, evictionPolicy, cfg.Cache.Size)
	return &Node{
		config: cfg,
		cache:  c,
	}
}

func (n *Node) Start() error {
	addr := n.config.Addr()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	n.listener = ln

	go n.acceptLoop()
	return nil
}

func (n *Node) acceptLoop() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			break
		}
		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	for {
		nBytes, err := conn.Read(buf)
		if err != nil {
			break
		}

		var req Request
		if err := json.Unmarshal(buf[:nBytes], &req); err != nil {
			n.sendResponse(conn, Response{Error: err.Error()})
			continue
		}

		n.processRequest(conn, req)
	}
}

func (n *Node) processRequest(conn net.Conn, req Request) {
	var resp Response

	switch req.Type {
	case "get":
		val, err := n.cache.Get(req.Key)
		if err != nil {
			resp.Error = err.Error()
		} else if val != nil {
			resp.Value = val
		}
	case "set":
		ttl := time.Duration(req.TTL) * time.Millisecond
		err := n.cache.Set(req.Key, req.Value, ttl)
		if err != nil {
			resp.Error = err.Error()
		}
	case "delete":
		resp.Flag = n.cache.Delete(req.Key)
	case "exist":
		resp.Flag = n.cache.Exist(req.Key)
	case "stat":
		resp.Stat = n.cache.Stat()
	case "ping":
		resp.Value = json.RawMessage(`"PONG"`)
	}

	n.sendResponse(conn, resp)
}

func (n *Node) sendResponse(conn net.Conn, resp Response) {
	data, _ := json.Marshal(resp)
	conn.Write(data)
}

func (n *Node) Stop() error {
	if n.listener != nil {
		return n.listener.Close()
	}
	return nil
}

type NodeClient struct {
	conn net.Conn
}

func NewNodeClient(addr string) (*NodeClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &NodeClient{conn: conn}, nil
}

func (c *NodeClient) Get(key string) (any, error) {
	req := Request{Type: "get", Key: key}
	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (c *NodeClient) Set(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	req := Request{Type: "set", Key: key, Value: data, TTL: int64(ttl.Milliseconds())}
	_, err = c.sendRequest(req)
	return err
}

func (c *NodeClient) Delete(key string) (bool, error) {
	req := Request{Type: "delete", Key: key}
	resp, err := c.sendRequest(req)
	if err != nil {
		return false, err
	}
	return resp.Flag, nil
}

func (c *NodeClient) Exist(key string) (bool, error) {
	req := Request{Type: "exist", Key: key}
	resp, err := c.sendRequest(req)
	if err != nil {
		return false, err
	}
	return resp.Flag, nil
}

func (c *NodeClient) Stat() (storage.Stat, error) {
	req := Request{Type: "stat"}
	resp, err := c.sendRequest(req)
	if err != nil {
		return storage.Stat{}, err
	}
	return resp.Stat, nil
}

func (c *NodeClient) sendRequest(req Request) (Response, error) {
	data, _ := json.Marshal(req)
	c.conn.Write(data)

	buf := make([]byte, 4096)
	n, err := c.conn.Read(buf)
	if err != nil {
		return Response{}, err
	}

	var resp Response
	json.Unmarshal(buf[:n], &resp)
	return resp, nil
}

func (c *NodeClient) Close() error {
	return c.conn.Close()
}

type DistributedCache struct {
	ring  *hash.ConsistentHashRing
	nodes map[string]*NodeClient
}

func NewDistributedCache(nodeConfigs []hash.NodeInfo) *DistributedCache {
	ring := hash.New(100)
	nodes := make(map[string]*NodeClient)

	for _, cfg := range nodeConfigs {
		ring.AddNode(hash.NodeInfo{
			ID:     cfg.ID,
			Addr:   cfg.Addr,
			Port:   cfg.Port,
			Weight: cfg.Weight,
		})

		addr := fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port)
		client, _ := NewNodeClient(addr)
		nodes[cfg.ID] = client
	}

	return &DistributedCache{
		ring:  ring,
		nodes: nodes,
	}
}

func (d *DistributedCache) Get(key string) (any, error) {
	nodeID := d.ring.GetNode(key)
	client := d.nodes[nodeID]
	if client == nil {
		return nil, fmt.Errorf("no node found for key: %s", key)
	}
	return client.Get(key)
}

func (d *DistributedCache) Set(key string, value any, ttl time.Duration) error {
	nodeID := d.ring.GetNode(key)
	client := d.nodes[nodeID]
	if client == nil {
		return fmt.Errorf("no node found for key: %s", key)
	}
	return client.Set(key, value, ttl)
}

func (d *DistributedCache) Delete(key string) (bool, error) {
	nodeID := d.ring.GetNode(key)
	client := d.nodes[nodeID]
	if client == nil {
		return false, fmt.Errorf("no node found for key: %s", key)
	}
	return client.Delete(key)
}

func (d *DistributedCache) Invalidate(key string) error {
	for _, client := range d.nodes {
		client.Delete(key)
	}
	return nil
}

func (d *DistributedCache) Stat() map[string]storage.Stat {
	stats := make(map[string]storage.Stat)
	for id, client := range d.nodes {
		stat, err := client.Stat()
		if err == nil {
			stats[id] = stat
		}
	}
	return stats
}
