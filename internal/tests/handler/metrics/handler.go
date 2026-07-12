// Package metrics — named type dependencies (Port, Timeout) и error-returning constructor (#28).
package metrics

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/handler"
	testLogger "github.com/diplexhq/diplex/internal/tests/logger"
	"github.com/diplexhq/diplex/internal/tests/metrics"
)

type Logger interface {
	testLogger.Logger
}

type Handler struct {
	handler.Base
	port    metrics.Port
	timeout metrics.Timeout
	checker *metrics.HealthChecker
	log     Logger
}

func New(
	port metrics.Port,
	timeout metrics.Timeout,
	checker *metrics.HealthChecker,
	log Logger,
) *Handler {
	return &Handler{
		Base:    handler.NewBase("/metrics"),
		port:    port,
		timeout: timeout,
		checker: checker,
		log:     log,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
