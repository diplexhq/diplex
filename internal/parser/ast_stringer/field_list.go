package ast_stringer

import (
	"go/ast"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
)

// FieldsToStrings converts AST field lists to a slice of parameter strings.
// Each struct/field parameter type is serialized independently.
// Returns []domain.Parameter directly, avoiding an intermediate []string conversion.
func FieldsToStrings(fields *ast.FieldList, imports map[string]string, pkg string, generics map[string]string) []domain.Parameter {
	if fields == nil || len(fields.List) == 0 {
		return nil
	}

	out := make([]domain.Parameter, 0, len(fields.List))

	s := astStringer{
		imports:  imports,
		generics: generics,
		pkg:      pkg,
		buf:      &strings.Builder{},
	}
	for _, f := range fields.List {
		s.buf.Reset()
		s.buf.Grow(128)
		s.writeExpr(f.Type)
		out = append(out, domain.Parameter(s.buf.String()))
	}

	return out
}
