package main

import (
	"os"

	"github.com/diplexhq/diplex/internal/generated/diplex"
)

// main is the entry point for diplex tool.
// Orchestrates scanning, parsing, and code generation phases.
func main() {
	deps := diplex.NewDI()
	log := deps.Logger()

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic recovered", "error", r)
			os.Exit(1)
		}
	}()

	// ── Dependencies ──
	cfg := deps.Config()
	scanner := deps.Scanner()
	parser := deps.Parser()
	res := deps.Resolver()
	gen := deps.Generator()
	// ── Verbose config logging ──
	log.Debug("config",
		"scan_dirs", cfg.ScanDirs(),
		"output_dir", cfg.OutputDir(),
		"skip_pattern", cfg.SkipPattern(),
		"module", cfg.Module(),
		"di_dirs", cfg.DIDirs())

	gen.Generate(res.Resolve(parser.Parse(scanner.Scan())))

	log.Info("generation complete")
}
