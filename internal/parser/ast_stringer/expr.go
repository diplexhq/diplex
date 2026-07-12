package ast_stringer

import (
	"fmt"
	"go/ast"
	"strings"
)

// ExprToString converts an AST expression node to its string representation.
// The aliases map resolves package aliases (e.g. "http" → "net/http") and
// prefixes unexported types with the current package namespace.
func ExprToString(expr ast.Expr, imports map[string]string, pkg string) string {
	return ExprToStringWithGenerics(expr, imports, pkg, nil)
}

// ExprToStringWithGenerics converts an AST expression node to its string representation,
// substituting generic type parameters using the provided generics map.
func ExprToStringWithGenerics(expr ast.Expr, imports map[string]string, pkg string, generics map[string]string) string {
	sb := &strings.Builder{}
	sb.Grow(128)
	s := astStringer{imports: imports, generics: generics, pkg: pkg, buf: sb}
	s.writeExpr(expr)

	return s.buf.String()
}

// writeExpr dispatches to the appropriate handler for each expression type.
func (s *astStringer) writeExpr(expr ast.Expr) {
	if expr == nil {
		return
	}

	switch expr := expr.(type) {
	case *ast.Ident:
		s.writeIdent(expr)
	case *ast.SelectorExpr:
		s.writeSelectorExpr(expr)
	case *ast.StarExpr:
		s.writeStarExpr(expr)
	case *ast.ArrayType:
		s.writeArrayType(expr)
	case *ast.MapType:
		s.writeMapType(expr)
	case *ast.BasicLit:
		s.writeBasicLit(expr)
	case *ast.StructType:
		s.writeStructType(expr)
	case *ast.InterfaceType:
		s.writeInterfaceType(expr)
	case *ast.FuncType:
		s.writeFuncType(expr)
	case *ast.ChanType:
		s.writeChanType(expr)
	case *ast.Ellipsis:
		s.writeEllipsis(expr)
	case *ast.IndexExpr:
		s.writeIndexExpr(expr)
	case *ast.IndexListExpr:
		s.writeIndexListExpr(expr)

	default:
		panic(fmt.Sprintf("unsupported AST expression type %T — extend ast_stringer or file issue", expr))
	}
}
