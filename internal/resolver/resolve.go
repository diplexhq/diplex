package resolver

import (
	"github.com/diplexhq/diplex/internal/domain"
)

func (res *Resolver) Resolve(parsedData domain.ParsedData) domain.ResolvedData {
	facades := res.resolveFacades(parsedData)

	return domain.ResolvedData{
		Facades:   facades,
		Providers: res.resolveProviders(parsedData, facades),
	}
}
