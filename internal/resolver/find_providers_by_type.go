package resolver

import (
	"maps"

	"github.com/diplexhq/diplex/internal/domain"
)

// findProvidersByType looks up providers by normalized Result type using the pre-built index.
// It normalizes arg the same way (all generic params → T), then for each candidate
// runs compareParams to confirm exact match with constraints.
func (res *Resolver) findProvidersByType(wanted domain.Parameter, index map[string][]*domain.Provider) []*domain.Provider {
	candidates := index[res.normalizeGenericParameter(string(wanted))]
	providers := make([]*domain.Provider, 0, len(candidates))

	for _, candidate := range candidates {
		constraints := maps.Clone(candidate.Generic)
		if constraints == nil {
			constraints = make(map[string][]string)
		}

		if res.compareParams(candidate.Result, wanted, constraints) {
			providers = append(providers, res.resolveProvider(candidate, constraints)...)
		}
	}

	return providers
}
