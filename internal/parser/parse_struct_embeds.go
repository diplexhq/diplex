package parser

import (
	"go/ast"

	"github.com/diplexhq/diplex/internal/domain"
	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
	"github.com/diplexhq/diplex/internal/utils"
)

// parseStructEmbeds scans type declarations for struct types with embedded
// fields (anonymous fields). For each struct with embeds, it records the
// embedded type names (with "*" prefix for pointer embeds) in the embeds map.
// The key is the fully qualified struct name (e.g. "pkg.MyStruct").
func (fp *Parser) parseStructEmbeds(genDecl *ast.GenDecl, imports map[string]string, pkg string, state *parseState) {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok || structType.Fields == nil {
			continue
		} else if !utils.IsExported(typeSpec.Name.Name) {
			continue
		}

		interfaceID := domain.InterfaceID(pkg + "." + typeSpec.Name.Name)

		var embeds []domain.InterfaceID

		for _, field := range structType.Fields.List {
			// Anonymous field — this is an embed.
			if len(field.Names) == 0 {
				fieldType := fp.unStar(field.Type)
				embeds = append(embeds, domain.InterfaceID(astStringer.ExprToString(fieldType, imports, pkg)))
			}
		}

		if len(embeds) > 0 {
			state.mu.Lock()
			state.embeds[interfaceID] = embeds
			state.mu.Unlock()
			fp.log.Debug("struct embed", "interface", interfaceID, "embeds", embeds)
		}
	}
}
