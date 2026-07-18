// Package di defines the application dependency injection facade.
// diplex generates the implementation for this interface.
package di

import (
	"github.com/diplexhq/diplex/internal/config"
	"github.com/diplexhq/diplex/internal/generator"
	"github.com/diplexhq/diplex/internal/parser"
	"github.com/diplexhq/diplex/internal/resolver"
	"github.com/diplexhq/diplex/internal/scanner"
	"github.com/diplexhq/diplex/internal/utils/logger"
)

// DI is the DI facade — all project dependencies.
type DI interface {
	Config() *config.Config
	Logger() logger.Logger
	Scanner() *scanner.Scanner
	Parser() *parser.Parser
	Resolver() *resolver.Resolver
	Generator() *generator.Generator
}
