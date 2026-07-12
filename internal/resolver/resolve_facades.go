package resolver

import (
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
)

func (res *Resolver) resolveFacades(parsedData domain.ParsedData) domain.Interfaces {
	facades := make(domain.Interfaces)

	for _, dir := range res.cfg.DIDirs() {
		base := res.cfg.Module() + "/" + dir
		for interfaceID, interfaceInfo := range parsedData.Interfaces {
			if rest, ok := strings.CutPrefix(string(interfaceID), base); ok {
				if len(rest) > 0 && (rest[0] == '.' || rest[0] == '/') {
					facades[interfaceID] = interfaceInfo
				}
			}
		}
	}

	return facades
}
