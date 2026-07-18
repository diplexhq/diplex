// Package usercreate — узкий интерфейс Create для теста narrow interface per consumer.
package create

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/entity"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type User = entity.User

type Repo interface {
	Set(string, User)
}

type Handler struct {
	handler.Base
	repo Repo
}

func New(userRepo Repo) *Handler {
	return &Handler{Base: handler.NewBase("/users"), repo: userRepo}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
