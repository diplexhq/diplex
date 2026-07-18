package storage

import (
	"github.com/diplexhq/diplex/internal/tests/config"
)

type DbConnection struct {
	dbDsn config.Dsn
}

func NewDbConnection(dbDsn config.Dsn) *DbConnection {
	return &DbConnection{
		dbDsn: dbDsn,
	}
}

type DbStorage[K string | int, V any] struct {
	data map[K]V
	conn *DbConnection
}

func NewDbStorage[V any, K string | int](
	conn *DbConnection,
) *DbStorage[K, V] {
	return &DbStorage[K, V]{
		conn: conn,
		data: make(map[K]V),
	}
}

func (s *DbStorage[K, V]) Get(key K) (V, bool) {
	val, ok := s.data[key]
	return val, ok
}

func (s *DbStorage[K, V]) Set(key K, val V) {
	s.data[key] = val
}

func (s *DbStorage[K, V]) Delete(key K) {
	delete(s.data, key)
}
