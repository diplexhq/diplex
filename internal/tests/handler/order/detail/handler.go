// Package orderdetail — multi-level alias unwinding: OrderDetail.Repo.Detail() → Order = entity.Order.
package detail

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Order = entity.Order

type Repo interface {
	Detail(id int) (Order, error)
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(repo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/orders/{id}/detail"), repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
