package resolver

import (
	"github.com/diplexhq/diplex/internal/domain"
)

// buildTypeIndex creates a map from normalized provider Result types to their pointers.
// All generic type parameters (T, T0, T1, ...) and all concrete types inside brackets
// are normalized to a single "T". Only non-generic providers (Result has no brackets)
// are also indexed by their exact Result.
func (res *Resolver) buildTypeIndex(parsedData domain.ParsedData) map[string][]*domain.Provider {
	index := make(map[string][]*domain.Provider)

	for _, provider := range parsedData.Providers {
		key := res.normalizeGenericParameter(string(provider.Result))
		index[key] = append(index[key], provider)
	}

	return index
}

// buildInterfaceIndex maps each method of a provider's Result type to the providers
// that implement it. Generic methods (T or T\d+ in params/results) use bare name keys;
// non-generic methods use "Name() result" or "Name(args) (r1, r2)" syntax.
func (res *Resolver) buildInterfaceIndex(parsedData domain.ParsedData) map[string][]*domain.Provider {
	index := make(map[string][]*domain.Provider)

	for _, provider := range parsedData.Providers {
		baseResult, _ := res.extractGenericPrototype(provider)
		if info, ok := parsedData.Interfaces[domain.InterfaceID(baseResult)]; ok {
			for methodName, contract := range info.Methods {
				key := res.methodKey(methodName, contract)
				index[key] = append(index[key], provider)
			}
		}
	}

	return index
}
