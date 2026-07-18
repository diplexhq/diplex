package storage

import "github.com/diplexhq/diplex/internal/tests/config"

type RedisConnection struct {
	redisDsn config.Dsn
}

func NewRedisConnection(redisDsn config.Dsn) *RedisConnection {
	return &RedisConnection{
		redisDsn: redisDsn,
	}
}

type RedisStorage[K string | int, V any] struct {
	conn *RedisConnection
	data map[K]V
}

func NewRedisStorage[V any, K string | int](conn *RedisConnection) *RedisStorage[K, V] {
	return &RedisStorage[K, V]{
		conn: conn,

		data: make(map[K]V),
	}
}

func (s *RedisStorage[K, V]) Get(key K) (V, bool) {
	val, ok := s.data[key]
	return val, ok
}

func (s *RedisStorage[K, V]) Set(key K, val V) {
	s.data[key] = val
}

func (s *RedisStorage[K, V]) Delete(key K) {
	delete(s.data, key)
}
