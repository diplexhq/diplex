package parser

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"

	"github.com/diplexhq/diplex/internal/utils/logger"
)

func TestParseImport_BlankAndDotSkipped(t *testing.T) {
	t.Parallel()

	// Import spec with:
	//   - blank import:  _ "fmt"
	//   - dot import:    . "strings"
	//   - normal import: "encoding/json"
	decl := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Name: &ast.Ident{Name: "_"},
				Path: &ast.BasicLit{Kind: token.STRING, Value: `"fmt"`},
			},
			&ast.ImportSpec{
				Name: &ast.Ident{Name: "."},
				Path: &ast.BasicLit{Kind: token.STRING, Value: `"strings"`},
			},
			&ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: `"encoding/json"`},
			},
		},
	}

	imports := make(map[string]string)

	fp := New(logger.Noop{}, nil)
	fp.parseImports(decl, imports)

	expected := map[string]string{"json": "encoding/json"}
	if !reflect.DeepEqual(imports, expected) {
		t.Errorf("imports = %v, want %v", imports, expected)
	}
}
