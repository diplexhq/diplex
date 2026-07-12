package resolver

import "github.com/diplexhq/diplex/internal/domain"

// Config interface defines methods needed by the Resolver from application configuration.
type Config interface {
	DIDirs() []string
	OutputDir() string
	Module() string
	ScanDirs() []string
}

type providerIndex struct {
	byType   map[string][]*domain.Provider
	byMethod map[string][]*domain.Provider
}

type Resolver struct {
	cfg Config
}

func New(cnf Config) *Resolver {
	return &Resolver{
		cfg: cnf,
	}
}
