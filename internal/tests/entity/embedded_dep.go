package entity

// Pingable — interface with exactly one implementation.
type Pingable interface {
	Ping() string
}

type pingImpl struct{}

func NewPingable() Pingable {
	return &pingImpl{}
}

func (p *pingImpl) Ping() string { return "pong" }

// Countable — interface with exactly one implementation.
type Countable interface {
	Count() int
}

type countImpl struct{}

func NewCountable() Countable {
	return &countImpl{}
}

func (c *countImpl) Count() int { return 1 }
