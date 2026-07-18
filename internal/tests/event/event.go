package event

import (
	"context"
)

type Event struct {
	Type    string
	Payload string
}

type NotifyChan chan Event

func NewNotifyChan() NotifyChan {
	return make(NotifyChan, 16)
}

// ComplexResult — named return types.
type ComplexResult struct {
	Count int
	Total float64
	Err   error
}

type PayloadEntry int

// ComplexEvent — chan direction, map, context, named return combined (#31, #32, #34).
type ComplexEvent interface {
	StreamOutputs(context.Context) chan<- Event
	StreamInputs(context.Context) <-chan ComplexResult
	Process(context.Context, []any, map[string][]int32) (map[uint8][]interface{}, []PayloadEntry)
}

type complexEventImpl struct{}

func NewComplexEvent() ComplexEvent {
	return &complexEventImpl{}
}

func (c *complexEventImpl) StreamOutputs(_ context.Context) chan<- Event { return nil }
func (c *complexEventImpl) StreamInputs(_ context.Context) <-chan ComplexResult {
	return nil
}

func (c *complexEventImpl) Process(_ context.Context, _ []any, _ map[string][]int32) (map[uint8][]interface{}, []PayloadEntry) {
	return nil, nil
}
