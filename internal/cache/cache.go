package cache

import (
	"context"
	"sync"
	"time"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
)

type cacheEntity struct {
	order domain.OrderWithInformation
	time  time.Time
}

type Cache struct {
	mu    sync.RWMutex
	cache map[uuid.UUID]cacheEntity
	ttl   time.Duration
	timer *time.Ticker
	done  chan struct{}
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		cache: make(map[uuid.UUID]cacheEntity),
		ttl:   ttl,
		done:  make(chan struct{}),
	}
}

func (c *Cache) Set(id uuid.UUID, order domain.OrderWithInformation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entity := cacheEntity{
		order: order,
		time:  time.Now(),
	}
	c.cache[id] = entity
}

func (c *Cache) Get(id uuid.UUID) (*domain.OrderWithInformation, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entity, ok := c.cache[id]
	if !ok {
		return nil, false
	}

	entity.time = time.Now()
	c.cache[id] = entity

	return &entity.order, true
}

func (c *Cache) StartDeleting(ctx context.Context) {
	c.timer = time.NewTicker(c.ttl / 2)

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.timer.Stop()
				return
			case <-c.done:
				c.timer.Stop()
				return
			case <-c.timer.C:
				c.mu.Lock()
				for id, entity := range c.cache {
					if time.Since(entity.time) > c.ttl {
						delete(c.cache, id)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
}

func (c *Cache) StopDeleting() {
	close(c.done)
}
