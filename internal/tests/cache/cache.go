// Package cache — generic cache с reversed type params [V, K] для тестирования order.
package cache

import "time"

type MemStorage[K string | int, V any] struct {
	data map[K]V
}

func NewMemStorage[V any, K string | int]() *MemStorage[K, V] {
	return &MemStorage[K, V]{
		data: make(map[K]V),
	}
}

func (c *MemStorage[K, V]) Get(key K) (V, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *MemStorage[K, V]) Set(key K, val V) {
	c.data[key] = val
}

func (c *MemStorage[K, V]) Delete(key K) {
	delete(c.data, key)
}

func (c *MemStorage[V, K]) Len() int {
	return len(c.data)
}

type CacheEntry[V any] struct {
	value V
	exp   time.Time
}
type Cache[K string | int, V any] struct {
	storage *MemStorage[K, CacheEntry[V]]
}

func New[K string | int, V any](storage *MemStorage[K, CacheEntry[V]]) *Cache[K, V] {
	return &Cache[K, V]{
		storage: storage,
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.storage.Get(key)

	return v.value, ok && v.exp.After(time.Now())
}

func (c *Cache[K, V]) Set(key K, val V, exp time.Time) {
	c.storage.Set(key, CacheEntry[V]{value: val, exp: exp})
}

func (c *Cache[K, V]) Delete(key K) {
	c.storage.Delete(key)
}

func (c *Cache[K, V]) Len() int {
	return c.storage.Len()
}
