package ast_stringer

import (
	"fmt"
	"go/ast"
	"sort"
)

// writeStarExpr writes "*X" for pointer types (e.g. *int, *MyType).
func (s *astStringer) writeStarExpr(expr *ast.StarExpr) {
	s.buf.WriteByte('*')
	s.writeExpr(expr.X)
}

// writeBasicLit writes a literal value (e.g. "int", "string", 42).
func (s *astStringer) writeBasicLit(lit *ast.BasicLit) {
	s.buf.WriteString(lit.Value)
}

// writeEllipsis writes "..." followed by the element type (variadic params).
func (s *astStringer) writeEllipsis(ellipsis *ast.Ellipsis) {
	s.buf.WriteString("...")
	s.writeExpr(ellipsis.Elt)
}

// writeStructType writes the struct type with its fields as a string.
// For empty structs: "struct{}"
// For structs with fields: "struct{ Host string; Port int }"
func (s *astStringer) writeStructType(st *ast.StructType) {
	s.buf.WriteString("struct{")

	if st.Fields != nil && st.Fields.List != nil {
		first := true
		for _, field := range st.Fields.List {
			if first {
				s.buf.WriteByte(' ')
			} else {
				s.buf.WriteString("; ")
			}

			first = false

			for _, name := range field.Names {
				s.buf.WriteString(name.Name)
				s.buf.WriteByte(' ')
			}

			s.writeExpr(field.Type)
		}

		s.buf.WriteByte(' ')
	}

	s.buf.WriteByte('}')
}

// writeInterfaceType writes an interface type with its methods and embedded
// interfaces. Methods are sorted alphabetically for canonical representation.
// For empty interfaces: "any". For interfaces with content:
// "interface{ Do(x int) string; EmbedInterface }".
func (s *astStringer) writeInterfaceType(it *ast.InterfaceType) {
	if it.Methods == nil || len(it.Methods.List) == 0 {
		s.buf.WriteString("any")
		return
	}

	// Sort methods by name for canonical representation.
	// Go considers interface method order irrelevant — { A(); B() } == { B(); A() }.
	methods := make([]*ast.Field, len(it.Methods.List))
	copy(methods, it.Methods.List)
	sort.Slice(methods, func(i, j int) bool {
		return s.methodSortKey(methods[i]) < s.methodSortKey(methods[j])
	})

	s.buf.WriteString("interface{")

	first := true
	for _, method := range methods {
		if first {
			s.buf.WriteByte(' ')
		} else {
			s.buf.WriteString("; ")
		}

		first = false

		switch mt := method.Type.(type) {
		case *ast.FuncType:
			// Method with signature: Name(params) results
			if len(method.Names) > 0 {
				s.buf.WriteString(method.Names[0].Name)
			}

			s.writeMethodSignature(mt)
		default:
			// Embedded interface — write raw name without namespace prefix
			s.writeEmbed(method.Type)
		}
	}

	s.buf.WriteString(" }")
}

// methodSortKey returns the sort key for an interface method field.
// For methods: the method name. For embeds: the type string representation.
func (s *astStringer) methodSortKey(field *ast.Field) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}
	// Embedded type — use its string representation
	return ExprToString(field.Type, s.imports, s.pkg)
}

// writeMethodSignature writes a function type as a method signature (no "func" keyword).
// Examples: "(x int) string", "(a, b int) (bool, error)", "()".
func (s *astStringer) writeMethodSignature(ft *ast.FuncType) {
	s.buf.WriteByte('(')

	s.writeFuncParams(ft)
}

func (s *astStringer) writeFuncParams(ft *ast.FuncType) {
	if ft.Params != nil && len(ft.Params.List) > 0 {
		s.writeFieldList(ft.Params)
	}

	s.buf.WriteByte(')')

	if ft.Results != nil && len(ft.Results.List) > 0 {
		if len(ft.Results.List) == 1 && len(ft.Results.List[0].Names) == 0 {
			// Single unnamed result: "(x int) string" (no parens)
			s.buf.WriteByte(' ')
			s.writeExpr(ft.Results.List[0].Type)
		} else {
			// Multiple results or named results: "(a, b int) (bool, error)"
			s.buf.WriteString(" (")
			s.writeFieldList(ft.Results)
			s.buf.WriteByte(')')
		}
	}
}

// writeFuncType writes a function type with parameters and results.
// Examples: "func(int) string", "func(a, b int) (bool, error)", "func()".
func (s *astStringer) writeFuncType(ft *ast.FuncType) {
	s.buf.WriteString("func(")
	s.writeFuncParams(ft)
}

// writeFieldList writes a comma-separated list of field types (params or results).
// Parameter NAMES are intentionally omitted — they don't affect type identity
// in Go and would break interface/implementation matching if different.
func (s *astStringer) writeFieldList(fl *ast.FieldList) {
	if fl == nil || len(fl.List) == 0 {
		return
	}

	first := true
	for _, field := range fl.List {
		if !first {
			s.buf.WriteString(", ")
		}

		first = false

		// Always write only the type, never the name.
		// Parameter names are irrelevant for type comparison in Go.
		s.writeExpr(field.Type)
	}
}

// writeEmbed writes an embedded interface name without namespace prefixing.
// Handles both Ident (EmbedInterface) and SelectorExpr (pkg.EmbedInterface).
func (s *astStringer) writeEmbed(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		s.buf.WriteString(e.Name)
	case *ast.SelectorExpr:
		// For qualified embeds, resolve the import alias
		s.writeExpr(e)
	default:
		panic(fmt.Sprintf("unsupported embed expression type %T — extend ast_stringer or file issue", expr))
	}
}

// writeChanType writes channel direction + type (e.g. "chan int", "<-chan int", "chan<- int").
func (s *astStringer) writeChanType(chanType *ast.ChanType) {
	switch chanType.Dir {
	case ast.SEND | ast.RECV:
		s.buf.WriteString("chan ")
	case ast.RECV:
		s.buf.WriteString("<-chan ")
	case ast.SEND:
		s.buf.WriteString("chan<- ")
	}

	s.writeExpr(chanType.Value)
}

// writeArrayType writes array/slice type (e.g. "[5]int", "[]byte").
// For slices, Len is nil and writeExpr skips it, producing "[]T".
func (s *astStringer) writeArrayType(arrayType *ast.ArrayType) {
	s.buf.WriteByte('[')
	s.writeExpr(arrayType.Len)
	s.buf.WriteByte(']')
	s.writeExpr(arrayType.Elt)
}

// writeMapType writes map type (e.g. "map[string]int").
func (s *astStringer) writeMapType(mapType *ast.MapType) {
	s.buf.WriteString("map[")
	s.writeExpr(mapType.Key)
	s.buf.WriteByte(']')
	s.writeExpr(mapType.Value)
}

// writeSelectorExpr writes qualified name (e.g. "pkg.TypeName").
func (s *astStringer) writeSelectorExpr(expr *ast.SelectorExpr) {
	s.writeExpr(expr.X)
	s.buf.WriteByte('.')
	s.buf.WriteString(expr.Sel.Name)
}

// writeIndexExpr writes a generic type with a single type argument
// (e.g. "Repo[Order]", "Handler[context.Context, string]").
func (s *astStringer) writeIndexExpr(ix *ast.IndexExpr) {
	s.writeExpr(ix.X)
	s.buf.WriteByte('[')
	s.writeExpr(ix.Index)
	s.buf.WriteByte(']')
}

// writeIndexListExpr writes a generic type with multiple type arguments
// (e.g. "Cache[string, SessionData]", "Map[K, V]").
func (s *astStringer) writeIndexListExpr(ix *ast.IndexListExpr) {
	s.writeExpr(ix.X)
	s.buf.WriteByte('[')

	for i, idx := range ix.Indices {
		if i > 0 {
			s.buf.WriteString(", ")
		}

		s.writeExpr(idx)
	}

	s.buf.WriteByte(']')
}
