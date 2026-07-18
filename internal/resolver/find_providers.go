package resolver

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

func (res *Resolver) findProviderCollection(arg domain.Parameter, parsedData domain.ParsedData, index providerIndex) (providers domain.ProviderCollection) {
	providers.Providers = res.findProviders(arg, parsedData, index)

	if len(providers.Providers) == 0 {
		if strings.HasPrefix(string(arg), "[]") {
			providers = res.findProviderCollection(domain.Parameter(string(arg)[2:]), parsedData, index)
			if providers.CollectionType != "" {
				panic(fmt.Sprintf("cannot generate DI: nested collection for %s is not supported", arg))
			}

			sort.Slice(providers.Providers, func(i, j int) bool {
				return providers.Providers[i].ID < providers.Providers[j].ID
			})

			return domain.ProviderCollection{
				CollectionType: "slice",
				Providers:      providers.Providers,
			}
		}

		panic(fmt.Sprintf("cannot generate DI: no providers for %s", arg))
	}

	return providers
}

// findProviders dispatches to type-only or interface-only search.
func (res *Resolver) findProviders(wanted domain.Parameter, parsedData domain.ParsedData, index providerIndex) []*domain.Provider {
	if info, ok := parsedData.Interfaces[domain.InterfaceID(wanted)]; ok && !info.RealType && len(info.Methods) > 0 {
		return res.findProvidersByInterface(wanted, parsedData, index.byMethod)
	}

	return res.findProvidersByType(wanted, index.byType)
}

// compareParams checks if providerArg satisfies interfaceArg by narrowing the constraints map.
// It compares type signatures character by character, resolving generic parameters (T, T0, etc.)
// against concrete types using the constraints map.
//
// Algorithm:
//  1. Exact match → true (no constraints needed).
//  2. Empty constraints with mismatch → false.
//  3. Loop: consume tokens from both strings.
//     - Non-generic token → exact match required (both sides advanced).
//     - Generic token on provider side → squeeze constraint with interface's concrete type.
//  4. Both strings must be fully consumed.
func (res *Resolver) compareParams(gotParam, wantParam domain.Parameter, constraints map[string][]string) bool {
	if gotParam == wantParam {
		return true
	} else if len(constraints) == 0 {
		return false
	}

	var wantToken, gotToken string

	got := string(gotParam)
	want := string(wantParam)

	for got != "" && want != "" {
		gotToken, got = utils.NextIdentToken(got)
		if !utils.IsGenericParam(gotToken) {
			wantToken, want = utils.NextIdentToken(want)
			if wantToken != gotToken {
				return false
			}

			continue
		}

		wantToken, want = utils.NextConcreteToken(want)
		if !res.squeezeConstraint(constraints, gotToken, wantToken) {
			return false
		}
	}

	return got == want
}

// squeezeConstraint validates concreteType against constraints[paramName] and narrows
// the constraint slice to contain only concreteType. Creates a new entry if key is missing.
// Returns false if concreteType is not in the allowed list.
func (res *Resolver) squeezeConstraint(constraints map[string][]string, paramName, concreteType string) bool {
	if existing, ok := constraints[paramName]; ok {
		if !slices.Contains(existing, concreteType) && !slices.Contains(existing, "any") {
			return false
		}

		constraints[paramName] = []string{concreteType}
	} else {
		constraints[paramName] = []string{concreteType}
	}

	return true
}

// resolveProvider builds concrete providers from a matching generic provider
// and its narrowed constraints. Returns all combination variants.
func (res *Resolver) resolveProvider(provider *domain.Provider, constraints map[string][]string) []*domain.Provider {
	if provider.Generic == nil {
		return []*domain.Provider{provider}
	}

	combinations := res.generateCombinations(constraints)
	providers := make([]*domain.Provider, 0, len(combinations))

	for _, combination := range combinations {
		result := utils.ReplaceTokens(string(provider.Result), combination)

		arguments := make([]domain.Parameter, 0, len(provider.Arguments))
		for _, a := range provider.Arguments {
			arguments = append(arguments, domain.Parameter(utils.ReplaceTokens(string(a), combination)))
		}

		name := utils.ReplaceTokens(provider.Name, combination)

		generic := make(map[string][]string)
		for template, constraint := range combination {
			generic[template] = []string{constraint}
		}

		providers = append(providers, &domain.Provider{
			ID:         domain.ProviderID(provider.Pkg + "." + name),
			Pkg:        provider.Pkg,
			Name:       name,
			Arguments:  arguments,
			ArgNames:   provider.ArgNames,
			Result:     domain.Parameter(result),
			ResultName: provider.ResultName,
			Generic:    nil,
			Error:      provider.Error,
		})
	}

	return providers
}

func (res *Resolver) generateCombinations(constraints map[string][]string) []map[string]string {
	keys := make([]string, 0, len(constraints))
	for k := range constraints {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var (
		result   []map[string]string
		generate func(idx int, current map[string]string)
	)

	generate = func(idx int, current map[string]string) {
		if idx == len(keys) {
			result = append(result, maps.Clone(current))

			return
		}

		k := keys[idx]

		opts := constraints[k]
		if len(opts) == 0 {
			generate(idx+1, current)
			return
		}

		for _, opt := range opts {
			current[k] = opt
			generate(idx+1, current)
		}
	}

	generate(0, make(map[string]string))

	return result
}
