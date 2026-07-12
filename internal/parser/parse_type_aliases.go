package parser

import (
	"go/ast"
	"go/token"

	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
)

// parseTypeAliases scans a Go file for type alias declarations (type X = Y)
// and populates the provided map with the fully qualified alias name
// (pkg.AliasName) and its underlying type.
func (fp *Parser) parseTypeAliases(genDecl *ast.GenDecl, imports map[string]string, pkg string, state *parseState) {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// A type alias has Assign != NoPos (the '=' token position).
		if typeSpec.Assign == token.NoPos {
			continue
		}

		name := typeSpec.Name.Name
		basic := astStringer.ExprToString(typeSpec.Type, imports, pkg)
		key := pkg + "." + name

		state.mu.Lock()
		state.typeAliases[key] = basic
		state.mu.Unlock()

		fp.log.Debug("type alias resolved", "from", key, "to", basic)
	}
}
