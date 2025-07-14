package cache

import (
	"sync"

	"github.com/SinevArtem/WBpage.git/internal/model"
)

type Cache struct {
	data map[string]model.Order
	mu   sync.Mutex
}

func New() *Cache {
	return &Cache{data: make(map[string]model.Order)}
}

func (c *Cache) Set(order model.Order) {
	c.mu.Lock()
	c.data[order.OrderUid] = order
	c.mu.Unlock()
}

func (c *Cache) Get(uid string) (model.Order, bool) {
	c.mu.Lock()
	order, found := c.data[uid]
	c.mu.Unlock()
	return order, found
}

func (c *Cache) GetAll() []model.Order {
	c.mu.Lock()
	all := make([]model.Order, 0, len(c.data))
	for _, i := range c.data {
		all = append(all, i)
	}
	c.mu.Unlock()

	return all
}
