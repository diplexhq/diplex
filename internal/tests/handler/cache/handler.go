// Package cache — generic cache с reversed type params для тестирования cache provider.
package cache

import (
	"net/http"
	"time"

	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Cache interface {
	Get(int) (string, bool)
	Set(int, string, time.Time)
}

type Handler struct {
	handler.Base
	cache Cache
}

func New(cache Cache) *Handler {
	return &Handler{
		Base:  handler.NewBase("/cache"),
		cache: cache,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
