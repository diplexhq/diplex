// Package orderdelete — узкий интерфейс для теста narrow interface per consumer.
package delete

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Repo interface {
	Delete(o entity.Order) error
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(repo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/orders/{id}"), repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
