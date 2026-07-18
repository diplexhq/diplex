package delete

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Repo interface {
	Delete(int)
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(orderRepo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/orders/{id}"), repo: orderRepo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
