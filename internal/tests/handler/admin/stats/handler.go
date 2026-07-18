package stats

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Repo interface {
	Stats() int
	Slug() string
}

type Handler struct {
	handler.Base
	repos []Repo
}

func New(repos []Repo) *Handler {
	return &Handler{Base: handler.NewBase("/admin/stats"), repos: repos}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
