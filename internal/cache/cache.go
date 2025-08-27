package cache

import (
	"sync"
	"web_demoservice/internal/model"
)

type Cache interface {
	Set(uid string, o model.Order)
	Get(uid string) (model.Order, bool)
}
type OrdersCache struct {
	mu   sync.RWMutex
	data map[string]model.Order
}

func NewOrdersCache() *OrdersCache {
	return &OrdersCache{data: make(map[string]model.Order)}
}

func (c *OrdersCache) Set(uid string, o model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[uid] = o
}

func (c *OrdersCache) Get(uid string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	o, ok := c.data[uid]
	return o, ok
}
