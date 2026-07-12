package parser

import (
	"go/ast"
	"path/filepath"
)

// parseImports extracts import declarations and populates the imports map.
// The key is the import name (explicit alias or last path segment),
// the value is the full import path without quotes.
func (fp *Parser) parseImports(genDecl *ast.GenDecl, imports map[string]string) {
	for _, spec := range genDecl.Specs {
		importSpec, ok := spec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		pkg := importSpec.Path.Value
		pkg = pkg[1 : len(pkg)-1] // strip surrounding quotes

		name := filepath.Base(pkg)
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		}

		// Skip blank imports (_ "pkg") — they have no usable alias.
		if name == "_" {
			continue
		}

		// Skip dot imports (. "pkg") — they pollute the current package
		// namespace and cannot be reliably resolved during alias matching.
		if name == "." {
			continue
		}

		imports[name] = pkg
	}
}
