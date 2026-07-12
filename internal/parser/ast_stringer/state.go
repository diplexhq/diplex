// Package ast_stringer converts Go AST expression nodes to their string
// representation. It is used by the parser to extract type signatures
// from constructors, interfaces, and implementations.
package ast_stringer

import "strings"

// astStringer holds state during a single conversion operation.
// It is an internal detail — callers use the package-level functions.
type astStringer struct {
	pkg      string            // fully qualified current package path
	imports  map[string]string // import alias map (key: alias name → value: import path)
	generics map[string]string // generic map
	buf      *strings.Builder  // output buffer, shared across all recursive calls
}
