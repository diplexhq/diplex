package delete

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type User = entity.User

type Repo interface {
	Delete(string)
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(userRepo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/users/{id}"), repo: userRepo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
