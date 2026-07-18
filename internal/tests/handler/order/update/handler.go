package update

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type OrderReader interface {
	Get(id int) (entity.Order, bool)
}

type OrderWriter interface {
	Set(int, entity.Order)
}

type Repo interface {
	OrderReader
	OrderWriter
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
