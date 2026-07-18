package storage

import "time"

type CacheEntry[V any] struct {
	value V
	exp   time.Time
}
type Cache[K string | int, V any] struct {
	storage *RedisStorage[K, CacheEntry[V]]
}

func New[K string | int, V any](storage *RedisStorage[K, CacheEntry[V]]) *Cache[K, V] {
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
