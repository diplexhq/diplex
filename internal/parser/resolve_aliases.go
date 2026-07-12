package parser

import (
	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

// resolveAliases flattens alias chains and replaces type alias references
// in providers, interfaces, and implementations with their underlying types.
func (fp *Parser) resolveAliases(state *parseState) {
	if len(state.typeAliases) == 0 {
		return
	}

	utils.ResolveReplacements(state.typeAliases)

	fp.resolveProviders(state)

	for _, info := range state.interfaces {
		fp.resolveMethods(state, info.Methods)
	}
}

func (fp *Parser) resolveProviders(state *parseState) {
	// Resolve both Arguments and Results, and re-key the map so that
	// aliased return types (e.g. *StringRepo → *Repo[string]) get the
	// correct concrete instantiation as their ServiceID.
	for providerID, provider := range state.providers {
		provider.Arguments = fp.resolveParams(state, provider.Arguments)
		provider.Result = domain.Parameter(fp.resolveTypeCached(state, string(provider.Result)))
		provider.Generic = fp.resolveConstraints(state, provider.Generic)
		state.providers[providerID] = provider
	}
}

// resolveParams replaces type aliases in each parameter with their underlying type.
func (fp *Parser) resolveParams(state *parseState, params []domain.Parameter) []domain.Parameter {
	if len(params) == 0 {
		return nil
	}

	resolved := make([]domain.Parameter, 0, len(params))
	for _, p := range params {
		resolved = append(resolved, domain.Parameter(fp.resolveTypeCached(state, string(p))))
	}

	return resolved
}

// resolveConstraints replaces type aliases in generic provider constraint strings.
// For example, if a constraint string contains "pkg.StringConstraint" and there is an
// alias "pkg.StringConstraint = int", it will be resolved to "int".
func (fp *Parser) resolveConstraints(state *parseState, generic map[string][]string) map[string][]string {
	if len(generic) == 0 {
		return generic
	}

	resolved := make(map[string][]string, len(generic))
	for param, constraints := range generic {
		if len(constraints) == 0 {
			resolved[param] = constraints
			continue
		}

		resolvedConstraints := make([]string, 0, len(constraints))
		for _, c := range constraints {
			resolvedConstraints = append(resolvedConstraints, fp.resolveTypeCached(state, c))
		}

		resolved[param] = resolvedConstraints
	}

	return resolved
}

// resolveMethods replaces type aliases in interface method signatures.
func (fp *Parser) resolveMethods(state *parseState, methods domain.MethodMap) {
	for name, method := range methods {
		method.Arguments = fp.resolveParams(state, method.Arguments)
		method.Results = fp.resolveParams(state, method.Results)
		methods[name] = method
	}
}

// resolveTypeCached replaces all alias occurrences and caches the result.
// Used during the resolve phase (providers, methods) to avoid re-computation.
func (fp *Parser) resolveTypeCached(state *parseState, t string) string {
	if resolved, ok := state.resolvedTypes[t]; ok {
		return resolved
	}

	result := utils.ReplaceTokens(t, state.typeAliases)
	state.resolvedTypes[t] = result

	return result
}
