package parser

import (
	"github.com/diplexhq/diplex/internal/domain"
)

// filterAutoProviders removes auto-provider entries that collide with explicit providers.
// An auto-provider is removed if an explicit provider (New* function) returns the same Result type.
func (fp *Parser) filterAutoProviders(state *parseState) {
	// Collect explicit provider Result types for O(1) collision lookup.
	explicitResults := make(map[domain.Parameter]struct{}, len(state.providers))

	for _, p := range state.providers {
		if !p.AutoProvider {
			explicitResults[p.Result] = struct{}{}
		}
	}

	// Remove auto-providers that collide with an explicit provider.
	for id, p := range state.providers {
		if !p.AutoProvider {
			continue
		}

		if _, collision := explicitResults[p.Result]; collision {
			delete(state.providers, id)
			fp.log.Debug("auto-provider skipped (explicit provider exists)", "id", id, "result", p.Result)
		}
	}
}
