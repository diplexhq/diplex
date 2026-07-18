package metrics

import (
	"errors"
	"time"

	"github.com/diplexhq/diplex/internal/tests/config"
)

type Port int

func NewPort() Port {
	return Port(8080)
}

type Timeout time.Duration

func NewTimeout() Timeout {
	return Timeout(30 * time.Second)
}

type HealthChecker struct {
	port Port
}

func NewHealthChecker(dbDsn, redisDsn config.Dsn, port Port) (*HealthChecker, error) {
	if dbDsn == "" {
		return nil, errors.New("metrics: db DSN is empty")
	}

	if redisDsn == "" {
		return nil, errors.New("metrics: redis DSN is empty")
	}

	return &HealthChecker{port: port}, nil
}

func (h *HealthChecker) Checker() Port {
	return h.port
}

type FactoryProvider struct {
	makeFunc HandlerFunc
}

type HandlerFunc func(name string) string

func (f *FactoryProvider) Make() HandlerFunc {
	return f.makeFunc
}
