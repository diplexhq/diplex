// Package metrics — named primitive types (Port, Timeout) и error-returning constructor для тестирования.
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

func NewHealthChecker(cfg config.DBConfig, port Port) (*HealthChecker, error) {
	if cfg.Dsn == "" {
		return nil, errors.New("metrics: DSN is empty")
	}

	return &HealthChecker{port: port}, nil
}

func (h *HealthChecker) Checker() Port {
	return h.port
}

type FactoryProvider struct {
	makeFunc HandlerFunc
}

func NewFactoryProvider(fn HandlerFunc) *FactoryProvider {
	return &FactoryProvider{makeFunc: fn}
}

type HandlerFunc func(name string) string

func NewHandlerFunc() HandlerFunc {
	return func(name string) string {
		return "factory:" + name
	}
}

func (f *FactoryProvider) Make() HandlerFunc {
	return f.makeFunc
}
