// Package orderupdate — interface composition via embedding (#15): Repo embeds OrderReader + OrderWriter.
package update

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type OrderReader interface {
	Get(id int) (entity.Order, error)
}

type OrderWriter interface {
	Update(o entity.Order) (entity.Order, error)
}

type Repo interface {
	OrderReader
	OrderWriter
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
