package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Option func(*Client)

type Client struct {
	mu      sync.RWMutex
	nodes   map[string]*nodeClient
	addr    string
	timeout time.Duration
}

type nodeClient struct {
	conn net.Conn
}

func New(addr string, opts ...Option) *Client {
	c := &Client{
		addr:    addr,
		timeout: 5 * time.Second,
		nodes:   make(map[string]*nodeClient),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

func (c *Client) getNode() (*nodeClient, error) {
	c.mu.RLock()
	nc := c.nodes[c.addr]
	c.mu.RUnlock()

	if nc != nil {
		return nc, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if nc, ok := c.nodes[c.addr]; ok {
		return nc, nil
	}

	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return nil, err
	}

	nc = &nodeClient{conn: conn}
	c.nodes[c.addr] = nc

	return nc, nil
}

type Request struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value any    `json:"value"`
	TTL   int64  `json:"ttl"`
}

type Response struct {
	Value any    `json:"value"`
	Flag  bool   `json:"flag"`
	Error string `json:"error,omitempty"`
}

func (c *Client) send(req Request) (*Response, error) {
	nc, err := c.getNode()
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(req)
	_, err = nc.conn.Write(data)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := nc.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	return &resp, nil
}

func (c *Client) Get(ctx context.Context, key string) (any, error) {
	req := Request{Type: "get", Key: key}
	resp, err := c.send(req)
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	req := Request{
		Type:  "set",
		Key:   key,
		Value: value,
		TTL:   int64(ttl.Milliseconds()),
	}
	_, err := c.send(req)
	return err
}

func (c *Client) Delete(ctx context.Context, key string) (bool, error) {
	req := Request{Type: "delete", Key: key}
	resp, err := c.send(req)
	if err != nil {
		return false, err
	}
	return resp.Flag, nil
}

func (c *Client) Exist(ctx context.Context, key string) (bool, error) {
	req := Request{Type: "exist", Key: key}
	resp, err := c.send(req)
	if err != nil {
		return false, err
	}
	return resp.Flag, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.send(Request{Type: "ping"})
	return err
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, nc := range c.nodes {
		nc.conn.Close()
	}
	c.nodes = make(map[string]*nodeClient)
	return nil
}

type ClusterClient struct {
	mu    sync.RWMutex
	nodes map[string]*Client
	ring  *ConsistentHashRing
}

func NewCluster(addrs []string, opts ...Option) *ClusterClient {
	ring := NewConsistentHashRing(100)
	nodes := make(map[string]*Client)

	for _, addr := range addrs {
		ring.AddNode(addr)
		nodes[addr] = New(addr, opts...)
	}

	return &ClusterClient{
		nodes: nodes,
		ring:  ring,
	}
}

func (c *ClusterClient) Get(ctx context.Context, key string) (any, error) {
	addr := c.ring.GetNode(key)
	client := c.nodes[addr]
	return client.Get(ctx, key)
}

func (c *ClusterClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	addr := c.ring.GetNode(key)
	client := c.nodes[addr]
	return client.Set(ctx, key, value, ttl)
}

func (c *ClusterClient) Delete(ctx context.Context, key string) (bool, error) {
	addr := c.ring.GetNode(key)
	client := c.nodes[addr]
	return client.Delete(ctx, key)
}

func (c *ClusterClient) Invalidate(ctx context.Context, key string) error {
	for _, client := range c.nodes {
		client.Delete(ctx, key)
	}
	return nil
}

func (c *ClusterClient) Close() error {
	for _, client := range c.nodes {
		client.Close()
	}
	return nil
}
