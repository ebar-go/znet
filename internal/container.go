package internal

import "sync"

// Container represents a safe map
type Container[Key int | string, Value any] struct {
	rmu   sync.RWMutex
	items map[Key]Value
}

// Set sets the value of a key in the container
func (c *Container[Key, Value]) Set(key Key, value Value) {
	c.rmu.Lock()
	c.items[key] = value
	c.rmu.Unlock()
}

// Get returns the value associated
func (c *Container[Key, Value]) Get(key Key) (item Value, exist bool) {
	c.rmu.RLock()
	item, exist = c.items[key]
	c.rmu.RUnlock()
	return
}

// Del removes the item from the container
func (c *Container[Key, Value]) Del(key Key) {
	c.rmu.Lock()
	delete(c.items, key)
	c.rmu.Unlock()
}

// NewContainer creates a new Container instance
func NewContainer[Key int | string, Value any]() *Container[Key, Value] {
	return &Container[Key, Value]{
		items: make(map[Key]Value),
	}
}