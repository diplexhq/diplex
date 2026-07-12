package generator

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

// Generate generates all DI container files from the dependency IR.
func (dig *DiGenerator) Generate(resolvedData domain.ResolvedData) {
	utils.NoErr(os.MkdirAll(dig.cfg.OutputDir(), 0o750))

	used := make(map[string]struct{})

	for _, facadeID := range dig.facadeIDs(resolvedData.Facades) {
		name := facadeID.LocalName()
		dig.render(
			dig.buildData(facadeID, resolvedData),
			dig.idUniq(name, used),
			dig.idUniq(utils.ToSnakeCase(name), used),
		)
	}
}

func (dig *DiGenerator) facadeIDs(facades domain.Interfaces) (facadeIDs []domain.InterfaceID) {
	facadeIDs = make([]domain.InterfaceID, 0, len(facades))
	for facadeID := range facades {
		facadeIDs = append(facadeIDs, facadeID)
	}

	slices.Sort(facadeIDs)

	return facadeIDs
}

func (dig *DiGenerator) idUniq(id string, used map[string]struct{}) string {
	variant := id
	for try := 0; ; try++ {
		if _, ok := used[variant]; !ok {
			break
		}

		variant = fmt.Sprintf("%s%d", variant, try)
	}

	used[variant] = struct{}{}

	return variant
}

func (dig *DiGenerator) simplePkg(pkg string) string {
	var ok bool

	alias := strings.TrimPrefix(pkg, dig.cfg.Module()+"/")
	for _, dir := range dig.cfg.ScanDirs() {
		if alias, ok = strings.CutPrefix(alias, dir+"/"); ok {
			break
		}
	}

	return utils.SanitizeIdent(alias)
}
