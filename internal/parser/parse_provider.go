package parser

import (
	"go/ast"
	"strconv"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
	astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
	"github.com/diplexhq/diplex/internal/utils"
)

// parseProvider extracts a provider contract from a function declaration.
// A valid provider:
//   - has no receiver (not a method),
//   - name starts with "New",
//   - returns exactly one type (pointer *T or value T),
//     optionally followed by error.
//
// Supports concrete generic instantiations: *Repo[Order], *Cache[string, V].
// Generic providers with type parameters (func NewRepo[T User | Order]()) are parsed
// with the type parameter aligned to "T" (constraints are stripped).
func (fp *Parser) parseProvider(funcDecl *ast.FuncDecl, imports map[string]string, pkg string, state *parseState) {
	if funcDecl.Recv != nil ||
		!strings.HasPrefix(funcDecl.Name.Name, "New") ||
		funcDecl.Type.Results == nil ||
		len(funcDecl.Type.Results.List) == 0 ||
		len(funcDecl.Type.Results.List) > 2 {
		return
	}

	genericAlias, genericSuffix, genericConstrain := fp.extractGenericParams(funcDecl, imports, pkg)

	res := fp.unStar(funcDecl.Type.Results.List[0].Type)

	baseIdent, ok := fp.baseIdent(res)
	if !ok || !utils.IsExported(baseIdent) {
		return
	}

	withError := false

	results, resultNames := astStringer.FieldsToStrings(funcDecl.Type.Results, imports, pkg, genericAlias)
	arguments, argNames := astStringer.FieldsToStrings(funcDecl.Type.Params, imports, pkg, genericAlias)

	resultName := baseIdent
	if len(resultNames) > 0 {
		resultName = resultNames[0]
	}

	if len(funcDecl.Type.Results.List) == 2 {
		errIdent, ok := funcDecl.Type.Results.List[1].Type.(*ast.Ident)
		if !ok || errIdent.Name != "error" {
			return
		}

		withError = true
	}

	provider := &domain.Provider{
		ID:         domain.ProviderID(pkg + "." + funcDecl.Name.Name + genericSuffix),
		Pkg:        pkg,
		Name:       funcDecl.Name.Name + genericSuffix,
		Arguments:  arguments,
		ArgNames:   argNames,
		Result:     results[0],
		ResultName: resultName,
		Generic:    genericConstrain,
		Error:      withError,
	}

	state.mu.Lock()
	state.providers[provider.ID] = provider
	state.mu.Unlock()

	fp.log.Debug("provider found", "id", provider.ID, "result", results[0])
}

// extractGenericParams extracts generic type parameters from a function declaration.
// Returns three values:
//   - Aliases: mapping from original param name → aligned name ("T", "T1", etc.)
//   - Suffix: bracketed list of aligned names for provider.Name (e.g. "[T]" or "[T, T1]")
//   - Constrains: mapping from aligned name → constraint strings (e.g. {"T": {"User", "Order"}} for [T User | Order])
//
// Returns zero values when no type parameters exist.
func (fp *Parser) extractGenericParams(funcDecl *ast.FuncDecl, imports map[string]string, pkg string) (map[string]string, string, map[string][]string) {
	if funcDecl.Type.TypeParams == nil || len(funcDecl.Type.TypeParams.List) == 0 {
		return nil, "", nil
	}

	aliases := make(map[string]string)
	constraints := make(map[string][]string)
	j := 0
	alias := "T"

	if len(funcDecl.Type.TypeParams.List) > 1 || len(funcDecl.Type.TypeParams.List[0].Names) > 1 {
		alias = "T0"
	}

	for _, field := range funcDecl.Type.TypeParams.List {
		for _, name := range field.Names {
			aliases[name.Name] = alias
			constraints[alias] = fp.constraintStrings(field.Type, imports, pkg)
			j++
			alias = "T" + strconv.FormatInt(int64(j), 10)
		}
	}

	return aliases, utils.TList(j), constraints
}

// constraintStrings converts a constraint expression into a slice of string representations.
// For union constraints like "User | Order", each operand is serialized separately.
// Simple constraints like "any" return a single-element slice.
func (fp *Parser) constraintStrings(expr ast.Expr, imports map[string]string, pkg string) []string {
	var result []string

	fp.walkConstraint(expr, func(e ast.Expr) {
		result = append(result, astStringer.ExprToStringWithGenerics(e, imports, pkg, nil))
	})

	return result
}

// walkConstraint visits each constraint expression, flattening union types (|).
func (fp *Parser) walkConstraint(expr ast.Expr, visit func(ast.Expr)) {
	if expr == nil {
		return
	}

	if be, ok := expr.(*ast.BinaryExpr); ok && be.Op.String() == "|" {
		fp.walkConstraint(be.X, visit)
		fp.walkConstraint(be.Y, visit)

		return
	}

	visit(expr)
}

// baseIdent extracts the base type name from an expression.
// Handles *ast.Ident (local types), *ast.SelectorExpr (qualified names like pkg.Type),
// *ast.IndexExpr (single type param like Repo[T]), and *ast.IndexListExpr (multiple type params).
// Returns ("", false) for unsupported expression types.
func (fp *Parser) baseIdent(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, true
	case *ast.SelectorExpr:
		return e.Sel.Name, true
	case *ast.IndexExpr:
		return fp.baseIdent(e.X)
	case *ast.IndexListExpr:
		return fp.baseIdent(e.X)
	default:
		return "", false
	}
}
