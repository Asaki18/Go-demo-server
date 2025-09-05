package cache

import (
	"encoding/json"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	cap  int
	data map[string]json.RawMessage
}

func New(limit int) *Cache {
	return &Cache{cap: limit, data: make(map[string]json.RawMessage, limit)}
}

func (c *Cache) Get(id string) (json.RawMessage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.data[id]
	return r, ok
}

func (c *Cache) Set(id string, raw json.RawMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cap > 0 && len(c.data) >= c.cap {
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}
	c.data[id] = raw
}

func (c *Cache) BulkLoad(m map[string]json.RawMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range m {
		if c.cap > 0 && len(c.data) >= c.cap {
			break
		}
		c.data[k] = v
	}
}