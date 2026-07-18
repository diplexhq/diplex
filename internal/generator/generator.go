package generator

import (
	"github.com/diplexhq/diplex/internal/utils/logger"
)

// Config interface defines methods needed by the Generator for code generation.
type Config interface {
	DIDirs() []string
	OutputDir() string
	Module() string
	ScanDirs() []string
}

// Generator produces DI container code from resolved dependency data.
type Generator struct {
	log logger.Logger
	cfg Config
}

// New creates a Generator with the given logger and config.
func New(log logger.Logger, cfg Config) *Generator {
	return &Generator{
		log: log,
		cfg: cfg,
	}
}
