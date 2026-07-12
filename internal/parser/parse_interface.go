package parser

import (
	"go/ast"

	"github.com/diplexhq/diplex/internal/domain"
	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
)

// parseInterface extracts interface definitions from a type declaration.
// It records method signatures and embedded interface names in the embeds map.
func (fp *Parser) parseInterface(genDecl *ast.GenDecl, imports map[string]string, pkg string, state *parseState) {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			continue
		}

		interfaceID := domain.InterfaceID(pkg + "." + typeSpec.Name.Name)
		methods := make(domain.MethodMap)

		var embeds []domain.InterfaceID

		for _, method := range interfaceType.Methods.List {
			switch mt := method.Type.(type) {
			case *ast.FuncType:
				if len(method.Names) == 0 {
					continue
				}

				methods[domain.FunctionName(method.Names[0].Name)] = domain.MethodContract{
					Arguments: astStringer.FieldsToStrings(mt.Params, imports, pkg, nil),
					Results:   astStringer.FieldsToStrings(mt.Results, imports, pkg, nil),
				}
			case *ast.Ident, *ast.SelectorExpr:
				embeds = append(embeds, domain.InterfaceID(astStringer.ExprToString(mt, imports, pkg)))
			}
		}

		state.mu.Lock()

		state.interfaces[interfaceID] = domain.InterfaceInfo{
			Methods:  methods,
			RealType: false, // declared interface, not a concrete type
		}
		if len(embeds) > 0 {
			state.embeds[interfaceID] = embeds
		}
		state.mu.Unlock()

		fp.log.Debug("interface parsed", "id", interfaceID, "methods", len(methods))
	}
}
