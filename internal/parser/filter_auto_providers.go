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
	for idx := 0; idx < len(state.providers); idx++ {
		if !state.providers[idx].AutoProvider {
			continue
		} else if _, collision := explicitResults[state.providers[idx].Result]; !collision {
			continue
		}

		fp.log.Debug("auto-provider skipped (explicit provider exists)", "id", state.providers[idx].ID, "result", state.providers[idx].Result)

		state.providers[idx] = state.providers[0]
		state.providers = state.providers[1:len(state.providers)]
		idx--
	}
}
