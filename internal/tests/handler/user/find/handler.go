package find

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type User = entity.User

type Repo interface {
	Find(args entity.FilterArgs) ([]User, error)
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(repo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/users"), repo: repo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
