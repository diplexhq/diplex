// Package orderget — узкий интерфейс Get(id int) для теста narrow interface per consumer.
package get

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Repo interface {
	Get(id int) (entity.Order, error)
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
