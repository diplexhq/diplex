// Package dispatcher — function type, channel type и ComplexEvent interface combined (#29, #30, #31).
package dispatcher

import (
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/callback"
	"github.com/diplexhq/diplex/internal/tests/event"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

type Handler struct {
	handler.Base
	fn      callback.HandlerFunc
	ch      event.NotifyChan
	complex event.ComplexEvent
}

func New(fn callback.HandlerFunc, ch event.NotifyChan, complex event.ComplexEvent) *Handler {
	return &Handler{
		Base:    handler.NewBase("/dispatch"),
		fn:      fn,
		ch:      ch,
		complex: complex,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ch <- event.Event{Type: "dispatch", Payload: h.fn("/dispatch")}

	_ = h.complex.Tags()

	w.WriteHeader(http.StatusOK)
}
