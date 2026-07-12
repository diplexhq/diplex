package logger

// Noop is a no-op Logger implementation for tests.
type Noop struct {
	Logger
}

// Debug does nothing.
func (Noop) Debug(_ string, _ ...any) {}

// Info does nothing.
func (Noop) Info(_ string, _ ...any) {}

// Error does nothing.
func (Noop) Error(_ string, _ ...any) {}
