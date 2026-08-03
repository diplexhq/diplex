package resolver

import (
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

	for _, facade := range facades {
		for _, method := range facade.Methods {
			if _, ok := argSeen[method.Results[0]]; !ok {
				queue = append(queue, method.Results[0])
				argSeen[method.Results[0]] = struct{}{}
			}
		}
	}

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
