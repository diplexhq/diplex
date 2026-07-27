package resolver

import (
	"maps"

	"github.com/diplexhq/diplex/internal/domain"
)

// findProvidersByInterface finds providers whose result struct satisfies the given interface
// by matching method signatures. Only called when arg is a real interface name.
func (res *Resolver) findProvidersByInterface(wanted domain.Parameter, parsedData domain.ParsedData, index map[string][]*domain.Provider) []*domain.Provider {
	wantedInfo := parsedData.Interfaces[domain.InterfaceID(wanted)]

	var fullKey, nameKey string

	for name, contract := range wantedInfo.Methods {
		fullKey = res.methodKey(name, contract)
		nameKey = string(name)

		break
	}

	candidates := index[fullKey]
	candidates = append(candidates, index[nameKey]...)

	if len(candidates) == 0 {
		return nil
	}

	providers := make([]*domain.Provider, 0, len(candidates))

	for _, candidate := range candidates {
		if candidate.Result == wanted {
			providers = append(providers, candidate)
			continue
		}

		baseResult, constraints := res.extractGenericPrototype(candidate)

		providerInfo, hasProviderInterface := parsedData.Interfaces[domain.InterfaceID(baseResult)]
		if !hasProviderInterface {
			continue
		}

		narrowTypeConstraints := maps.Clone(constraints)
		if candidate.Generic != nil {
			for k, v := range constraints {
				if _, ok := candidate.Generic[v[0]]; ok {
					narrowTypeConstraints[k] = candidate.Generic[v[0]]
				}
			}
		}

		if !res.methodMatches(wantedInfo, providerInfo, narrowTypeConstraints) {
			continue
		}

		providerConstraints := maps.Clone(candidate.Generic)
		if candidate.Generic != nil {
			for k, v := range constraints {
				if _, ok := providerConstraints[v[0]]; ok {
					providerConstraints[v[0]] = narrowTypeConstraints[k]
				}
			}
		}

		providers = append(providers, res.resolveProvider(candidate, providerConstraints)...)
	}

	return dedupAutoProviders(providers)
}

// dedupAutoProviders removes duplicate auto-provider variants for the same base type,
// preferring the pointer variant over the value variant when both exist.
func dedupAutoProviders(providers []*domain.Provider) []*domain.Provider {
	byType := make(map[string]*domain.Provider)

	for _, p := range providers {
		if !p.AutoProvider {
			continue
		}

		baseType := string(p.Result)
		if len(baseType) > 0 && baseType[0] == '*' {
			baseType = baseType[1:]
		}

		existing, ok := byType[baseType]
		if !ok {
			byType[baseType] = p
			continue
		}
		// Prefer pointer over value.
		if len(existing.Result) > 0 && existing.Result[0] == '*' {
			continue // existing is pointer, keep it
		}

		byType[baseType] = p // replace value with pointer
	}

	result := make([]*domain.Provider, 0, len(providers))
	for _, p := range providers {
		if !p.AutoProvider {
			result = append(result, p)
			continue
		}

		baseType := string(p.Result)
		if len(baseType) > 0 && baseType[0] == '*' {
			baseType = baseType[1:]
		}

		if byType[baseType] == p {
			result = append(result, p)
		}
	}

	return result
}

// methodMatches checks if all methods of the provider interface match the needed interface.
func (res *Resolver) methodMatches(wantedInfo, providerInfo domain.InterfaceInfo, constraints map[string][]string) bool {
	for name, contract := range wantedInfo.Methods {
		providerMethod, ok := providerInfo.Methods[name]
		if !ok {
			return false
		}

		if len(providerMethod.Arguments) != len(contract.Arguments) ||
			len(providerMethod.Results) != len(contract.Results) {
			return false
		}

		for i := range providerMethod.Arguments {
			if !res.compareParams(providerMethod.Arguments[i], contract.Arguments[i], constraints) {
				return false
			}
		}

		for i := range providerMethod.Results {
			if !res.compareParams(providerMethod.Results[i], contract.Results[i], constraints) {
				return false
			}
		}
	}

	return true
}
