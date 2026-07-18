package resolver

import (
	"fmt"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

func (res *Resolver) Resolve(parsedData domain.ParsedData) domain.ResolvedData {
	facades := res.resolveFacades(parsedData)
	providers := res.resolveProviders(parsedData, facades)

	visited := make(map[domain.ProviderID]struct{})
	queue := make([]*domain.Provider, 0, len(parsedData.Providers))
	resolvedFacades := make(map[domain.InterfaceID]domain.ResolvedFacade)
	resolvedProviders := make(map[domain.ProviderID]domain.ResolvedProvider)

	for interfaceID, facade := range facades {
		resolvedFacade := make(domain.ResolvedFacade)

		for functionName, methodContract := range facade.Methods {
			result := methodContract.Results[0]

			provider := res.selectSingleProvider(providers[result], "")
			if provider == nil {
				panic(fmt.Sprintf("no single provider found for %s:%s", interfaceID, functionName))
			}

			resolvedFacade[functionName] = domain.ResolvedFacadeMethod{
				Result:   result,
				Provider: provider,
			}

			if _, ok := visited[provider.ID]; ok {
				continue
			}

			queue = append(queue, provider)
			visited[provider.ID] = struct{}{}
		}

		resolvedFacades[interfaceID] = resolvedFacade
	}

	for len(queue) > 0 {
		provider := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		resolvedProvider := domain.ResolvedProvider{
			Provider:          provider,
			ArgumentProviders: make([]domain.ProviderCollection, 0, len(provider.Arguments)),
		}
		for i, arg := range provider.Arguments {
			providerCollection := providers[arg]
			if providerCollection.CollectionType != "" {
				resolvedProvider.ArgumentProviders = append(resolvedProvider.ArgumentProviders, providerCollection)
				continue
			}

			argumentProvider := res.selectSingleProvider(providerCollection, provider.ArgNames[i])
			if argumentProvider == nil {
				panic(fmt.Sprintf("no single provider found for %s(%s)", arg, provider.ArgNames[i]))
			}

			resolvedProvider.ArgumentProviders = append(resolvedProvider.ArgumentProviders, domain.ProviderCollection{
				Providers: []*domain.Provider{argumentProvider},
			})
		}

		for _, collection := range resolvedProvider.ArgumentProviders {
			for _, p := range collection.Providers {
				if _, ok := visited[p.ID]; ok {
					continue
				}

				queue = append(queue, p)
				visited[p.ID] = struct{}{}
			}
		}

		resolvedProviders[provider.ID] = resolvedProvider
	}

	return domain.ResolvedData{
		ResolvedFacades:   resolvedFacades,
		ResolvedProviders: resolvedProviders,
	}
}

// selectSingleProvider picks one provider from a multi-provider collection by
// matching provider ResultName against a target name.
//
// Returns nil if argName is empty or no unique name match is found — caller panics.
func (res *Resolver) selectSingleProvider(collection domain.ProviderCollection, argName string) *domain.Provider {
	if len(collection.Providers) == 1 {
		return collection.Providers[0]
	}

	if argName == "" {
		return nil
	}

	var matched []*domain.Provider

	for _, p := range collection.Providers {
		if res.matchByName(p.ResultName, argName) {
			matched = append(matched, p)
		}
	}

	if len(matched) != 1 {
		return nil
	}

	return matched[0]
}

// matchByName checks if provider's ResultName matches targetName.
func (res *Resolver) matchByName(resultName, argName string) bool {
	if len(resultName) == 0 || len(argName) == 0 ||
		len(resultName) != len(argName) {
		return false
	}

	return utils.ToLower(resultName[0]) == utils.ToLower(argName[0]) && resultName[1:] == argName[1:]
}
