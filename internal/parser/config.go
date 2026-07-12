package parser

// Config is what Processor needs from the application config.
// Declared locally — each consumer defines exactly what it needs.
type Config interface {
	Module() string
}
