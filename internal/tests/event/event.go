// Package event — chan direction (#23), map types (#24), named return values (#34) для тестирования.
package event

import "context"

type Event struct {
	Type    string
	Payload string
}

type NotifyChan chan Event

func NewNotifyChan() NotifyChan {
	return make(NotifyChan, 16)
}

// ComplexResult — named return types (#34).
type ComplexResult struct {
	Count int
	Total float64
	Err   error
}

// ComplexEvent — chan direction, map, context, named return combined (#31, #32, #34).
type ComplexEvent interface {
	StreamOutputs(ctx context.Context) chan<- Event
	StreamInputs(ctx context.Context) <-chan ComplexResult
	Hub() chan chan Event
	Tags() map[string]string
	Sizes() map[string]int
	Register(ctx context.Context, name string, handler chan Event) error
}

type complexEventImpl struct{}

func NewComplexEvent() ComplexEvent {
	return &complexEventImpl{}
}

func (c *complexEventImpl) StreamOutputs(ctx context.Context) chan<- Event        { return nil }
func (c *complexEventImpl) StreamInputs(ctx context.Context) <-chan ComplexResult { return nil }
func (c *complexEventImpl) Hub() chan chan Event                                  { return nil }
func (c *complexEventImpl) Tags() map[string]string                               { return nil }
func (c *complexEventImpl) Sizes() map[string]int                                 { return nil }
func (c *complexEventImpl) Register(ctx context.Context, name string, handler chan Event) error {
	return nil
}
