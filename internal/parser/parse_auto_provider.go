package parser

import (
	"go/ast"

	"github.com/diplexhq/diplex/internal/domain"
	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
	"github.com/diplexhq/diplex/internal/utils"
)

// parseAutoProvider registers both pointer and value providers for every exported struct.
// Embedded interfaces and structs are collected as dependencies.
// Both pointer (*StructName) and value (StructName) providers are generated.
//
// The pointer provider's arguments use the struct literal key: type name for embedded fields,
// field name for named fields. The value provider shares the same arguments but produces
// a value-typed Result (StructName, not *StructName).
func (fp *Parser) parseAutoProvider(typeSpec *ast.TypeSpec, structType *ast.StructType,
	imports map[string]string, pkg string, state *parseState,
) {
	// Collect all exported field types and names.
	fields := make([]domain.Parameter, 0, len(structType.Fields.List))
	argNames := make([]string, 0, len(structType.Fields.List))

	// First pass: named exported fields.
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if !utils.IsExported(name.Name) {
				continue
			}

			typeStr := astStringer.ExprToString(field.Type, imports, pkg)

			argNames = append(argNames, name.Name)
			fields = append(fields, domain.Parameter(typeStr))
		}
	}

	// Second pass: embedded (anonymous) fields.
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			continue // skip named fields
		}

		inner := fp.unStar(field.Type)

		localName := localTypeName(inner)
		if localName == "" {
			continue
		}

		if !utils.IsExported(localName) {
			continue
		}

		typeStr := astStringer.ExprToString(field.Type, imports, pkg)

		fields = append(fields, domain.Parameter(typeStr))
		argNames = append(argNames, localName)
	}

	// Always register auto-provider for exported structs (even empty ones like BaseStats).

	// Store fields for post-processing (struct embeds may collide with explicit providers).
	// At this point we don't know which embedded types are interfaces vs structs,
	// so we collect all and filter collisions in filterAutoProviders.

	// Create a synthetic provider: Result is pointer to struct.
	typeName := pkg + "." + typeSpec.Name.Name
	pointerType := "*" + typeName

	provider := &domain.Provider{
		ID:           domain.ProviderID(pointerType),
		Pkg:          pkg,
		Name:         typeSpec.Name.Name,
		Arguments:    fields,
		ArgNames:     argNames,
		Result:       domain.Parameter(pointerType),
		ResultName:   typeSpec.Name.Name,
		Generic:      nil,
		Error:        false,
		AutoProvider: true,
	}

	// Value provider (non-pointer variant).
	valueProvider := &domain.Provider{
		ID:           domain.ProviderID(typeName),
		Pkg:          pkg,
		Name:         typeSpec.Name.Name,
		Arguments:    fields,
		ArgNames:     argNames,
		Result:       domain.Parameter(typeName),
		ResultName:   typeSpec.Name.Name,
		Generic:      nil,
		Error:        false,
		AutoProvider: true,
	}

	state.mu.Lock()
	state.providers = append(state.providers, provider, valueProvider)
	state.mu.Unlock()

	fp.log.Debug("auto-provider found", "id", valueProvider, "fields", len(fields))
}

// localTypeName extracts the local (unqualified) type name from an AST expression.
// Returns "" if the type name cannot be determined.
func localTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}
