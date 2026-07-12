package generator

import (
	"github.com/diplexhq/diplex/internal/utils/logger"
)

// Config interface defines methods needed by the DiGenerator for code generation.
type Config interface {
	DIDirs() []string
	OutputDir() string
	Module() string
	ScanDirs() []string
}

// DiGenerator produces DI container code from resolved dependency data.
type DiGenerator struct {
	log logger.Logger
	cfg Config
}

// New creates a DiGenerator with the given logger and config.
func New(log logger.Logger, cfg Config) *DiGenerator {
	return &DiGenerator{
		log: log,
		cfg: cfg,
	}
}
