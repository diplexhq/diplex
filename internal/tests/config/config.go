// Package config — root provider (без deps) и named primitive для тестирования.
package config

// DBConfig — value type provider для тестирования.
type (
	Dsn      string
	RedisDsn = Dsn
)

// Config — корневой провайдер (без зависимостей).
type Config struct {
	db       Dsn
	redis    Dsn
	memcache Dsn
}

func NewConfig() *Config {
	return &Config{
		db:       "postgres://user:pass@localhost:5432/",
		redis:    "redis://user:pass@localhost:5432/",
		memcache: "memcache://user:pass@localhost:5432/",
	}
}

func NewRedisDsn(cnf *Config) RedisDsn {
	return cnf.redis
}

func NewDbDsn(cnf *Config) (dbDsn Dsn) {
	return cnf.db
}
