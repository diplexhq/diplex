package dispatcher

import (
	"context"
	"net/http"

	"github.com/diplexhq/diplex/internal/tests/event"
	"github.com/diplexhq/diplex/internal/tests/handler"
)

// ComplexEventWithComposite — narrow interface at point of use with any and rune composite types.
type ComplexEventWithComposite interface {
	StreamOutputs(context.Context) chan<- event.Event
	StreamInputs(context.Context) <-chan event.ComplexResult
	Process(context.Context, []any, map[string][]rune) (map[byte][]any, []event.PayloadEntry)
}

type Handler struct {
	handler.Base
	fn      event.HandlerFunc
	ch      event.NotifyChan
	complex ComplexEventWithComposite
}

func New(fn event.HandlerFunc, ch event.NotifyChan, complex ComplexEventWithComposite) *Handler {
	return &Handler{
		Base:    handler.NewBase("/dispatch"),
		fn:      fn,
		ch:      ch,
		complex: complex,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ch <- event.Event{Type: "dispatch", Payload: h.fn("/dispatch")}

	w.WriteHeader(http.StatusOK)
}
