package ast_stringer

import "go/ast"

// builtinTypes contains Go predeclared types that should NOT be prefixed with
// the package namespace. E.g., "string" stays "string", but "MyType" becomes "mypkg.MyType".
//
// NOTE: byte and uint8 (rune and int32) are the same type in Go, but the AST
// represents them as distinct identifiers ("byte" vs "uint8"). Since type
// matching in the DI generator uses string comparison, mixing aliases in
// interface declarations and implementations will cause a mismatch:
//
//	interface Foo { Bar(data byte) }   // Arguments: "byte"
//	func (s *S) Bar(data uint8) { }    // Arguments: "uint8"  → NO MATCH
//
// Workaround: use the same spelling consistently (prefer uint8/int32 over byte/rune).
var builtinTypes = map[string]struct{}{
	"any":        {},
	"bool":       {},
	"byte":       {}, // alias for uint8, but appears as a distinct identifier in AST
	"complex64":  {},
	"complex128": {},
	"comparable": {}, // built-in constraint, not a type — must stay unprefixed
	"error":      {},
	"float32":    {},
	"float64":    {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"rune":       {}, // alias for int32, but appears as a distinct identifier in AST
	"string":     {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
}

// writeIdent writes an identifier with namespace resolution.
//
// Resolution order:
//  1. If the name is in aliases (package alias map) → write the aliased name
//  2. If the name is a built-in type → write as-is
//  3. Otherwise → prefix with current package namespace
func (s *astStringer) writeIdent(ident *ast.Ident) {
	name := ident.Name
	if alias, ok := s.imports[name]; ok {
		name = alias
	} else if alias, ok = s.generics[name]; ok {
		name = alias
	} else if _, ok = builtinTypes[name]; !ok {
		s.buf.WriteString(s.pkg)
		s.buf.WriteByte('.')
	}

	s.buf.WriteString(name)
}
