package ast_stringer

import (
	"go/ast"
	"strings"

	"github.com/diplexhq/diplex/internal/domain"
)

// FieldsToStrings converts AST field lists to a slice of parameter strings.
// Each struct/field parameter type is serialized independently.
// Returns []domain.Parameter directly, avoiding an intermediate []string conversion.
func FieldsToStrings(
	fields *ast.FieldList,
	imports map[string]string,
	pkg string,
	generics map[string]string,
) (params []domain.Parameter, paramNames []string) {
	if fields == nil || len(fields.List) == 0 {
		return nil, nil
	}

	params = make([]domain.Parameter, 0, len(fields.List))
	paramNames = make([]string, 0, len(fields.List))

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
		t := s.buf.String()

		for _, name := range f.Names {
			paramNames = append(paramNames, name.Name)
			params = append(params, domain.Parameter(t))
		}

		if len(f.Names) == 0 {
			params = append(params, domain.Parameter(t))
		}
	}

	return params, paramNames
}
