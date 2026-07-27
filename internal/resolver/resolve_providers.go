package resolver

import (
	"sort"

	"github.com/diplexhq/diplex/internal/domain"
)

func (res *Resolver) resolveProviders(parsedData domain.ParsedData, facades domain.Interfaces) map[domain.Parameter]domain.ProviderCollection {
	index := providerIndex{
		byType:   res.buildTypeIndex(parsedData),
		byMethod: res.buildInterfaceIndex(parsedData),
	}

	argSeen := make(map[domain.Parameter]struct{})
	providers := make(map[domain.Parameter]domain.ProviderCollection)

	queue := make([]domain.Parameter, 0, 64)

	// Sort facade keys for deterministic queue building.
	facadeKeys := make([]string, 0, len(facades))
	for id := range facades {
		facadeKeys = append(facadeKeys, string(id))
	}

	sort.Strings(facadeKeys)

	for _, facadeID := range facadeKeys {
		facade := facades[domain.InterfaceID(facadeID)]
		// Sort method keys for deterministic order.
		methodKeys := make([]domain.FunctionName, 0, len(facade.Methods))
		for name := range facade.Methods {
			methodKeys = append(methodKeys, name)
		}

		sort.Slice(methodKeys, func(i, j int) bool { return methodKeys[i] < methodKeys[j] })

		for _, methodName := range methodKeys {
			method := facade.Methods[domain.FunctionName(methodName)]
			if _, ok := argSeen[method.Results[0]]; !ok {
				queue = append(queue, method.Results[0])
				argSeen[method.Results[0]] = struct{}{}
			}
		}
	}

	// Sort the queue for deterministic processing order.
	sort.SliceStable(queue, func(i, j int) bool {
		return queue[i] < queue[j]
	})

	for len(queue) > 0 {
		arg := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		collection := res.findProviderCollection(arg, parsedData, index)

		providers[arg] = collection
		for _, p := range collection.Providers {
			for _, a := range p.Arguments {
				if _, ok := argSeen[a]; ok {
					continue
				}

				argSeen[a] = struct{}{}

				queue = append(queue, a)
			}
		}
	}

	return providers
}
