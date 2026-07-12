package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils"
)

// parseFile parses a single Go source file and populates the provided AnalysisResult
// with constructors, interfaces, and implementations found within.
func (fp *Parser) parseFile(
	sourceFile domain.SourceFile,
	state *parseState,
) {
	node := utils.Must(parser.ParseFile(state.fSet, string(sourceFile), nil, parser.AllErrors))

	// imports maps import alias names to their resolved paths.
	imports := make(map[string]string)
	pkg := fp.cfg.Module() + "/" + filepath.Dir(string(sourceFile))

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				fp.parseImports(d, imports)
			case token.TYPE:
				fp.parseInterface(d, imports, pkg, state)
				fp.parseTypeAliases(d, imports, pkg, state)
				fp.parseStructEmbeds(d, imports, pkg, state)
			}
		case *ast.FuncDecl:
			fp.parseReceiver(d, imports, pkg, state)
			fp.parseProvider(d, imports, pkg, state)
		}
	}
}
