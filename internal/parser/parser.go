// Package parser parses Go source files and extracts constructors, interfaces,
// and implementations for DI container generation.
package parser

import (
	"go/ast"
	"go/token"
	"sync"

	"github.com/diplexhq/diplex/internal/domain"
	"github.com/diplexhq/diplex/internal/utils/logger"
)

// Parser analyses Go source files and extracts constructor functions, interfaces, and method implementations.
type Parser struct {
	log logger.Logger
	cfg Config
}

// New creates a Parser with the given logger and config.
func New(log logger.Logger, cfg Config) *Parser {
	return &Parser{
		log: log,
		cfg: cfg,
	}
}

// parseState holds intermediate data during file processing.
// It encapsulates the maps used for type alias resolution, embed tracking,
// and interface/provider collection. The mu field protects providers, interfaces,
// and embeds for concurrent access across goroutines.
type parseState struct {
	providers     domain.Providers
	interfaces    domain.Interfaces
	embeds        map[domain.InterfaceID][]domain.InterfaceID
	typeAliases   map[string]string
	resolvedTypes map[string]string
	mu            sync.Mutex
	fSet          *token.FileSet
}

// newState creates a new parseState with initialized maps and default type aliases.
func (fp *Parser) newState() *parseState {
	return &parseState{
		providers:  make(domain.Providers),
		interfaces: make(domain.Interfaces),
		embeds:     make(map[domain.InterfaceID][]domain.InterfaceID),
		typeAliases: map[string]string{
			"byte":        "uint8",
			"rune":        "int32",
			"interface{}": "any",
		},
		resolvedTypes: make(map[string]string),
		fSet:          token.NewFileSet(),
	}
}

// Parse analyzes files from the given channel and returns the collected metadata.
// Uses module from the injected Config to qualify internal import paths.
func (fp *Parser) Parse(sourceFiles domain.SourceFiles) domain.ParsedData {
	state := fp.newState()

	var g sync.WaitGroup

	for range 4 {
		g.Go(func() {
			for sourceFile := range sourceFiles {
				fp.log.Debug("parsing source", "file", sourceFile)
				fp.parseFile(sourceFile, state)
			}
		})
	}

	g.Wait()

	fp.resolveAliases(state)
	fp.resolveEmbeds(state)

	fp.log.Debug("analysis complete", "providers", len(state.providers), "interfaces", len(state.interfaces))

	return domain.ParsedData{
		Providers:  state.providers,
		Interfaces: state.interfaces,
	}
}

// unStar unwraps a pointer expression.
// Returns the underlying type for pointer expressions (*X → X), otherwise returns expr unchanged.
func (fp *Parser) unStar(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.StarExpr:
		return expr.X
	default:
		return expr
	}
}
