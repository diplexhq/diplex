// Package config — root provider (без deps) и named primitive для тестирования.
package config

// DBConfig — value type provider для тестирования.
type DBConfig struct {
	Dsn string
}

func (c DBConfig) DSN() string       { return c.Dsn }
func (c DBConfig) DSNMethod() string { return c.DSN() }

// Config — корневой провайдер (без зависимостей).
type Config struct {
	db DBConfig
}

func NewConfig() *Config {
	return &Config{
		db: DBConfig{Dsn: "postgres://user:pass@localhost:5432/" + "db"},
	}
}

func (c *Config) DB() DBConfig         { return c.db }
func NewDBConfig(cfg *Config) DBConfig { return cfg.db }
