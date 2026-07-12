# Development Guide

## Core Principles

- **Zero dependencies** — standard library only
- **Resource efficiency** — benchmarks and performance optimization
- **Self-applicable** — DIplex generates DI container for itself (self-hosting)
- **Test coverage** — all features implemented in `internal/tests/`, primary test coverage lives there
- **Minimalism** — don't overcomplicate without necessity
- **Performance** — any overhead must be justified
- **Type-safety** — check everything possible at compile-time
- **Readability** — generated code must be understandable

## CI / Git Hooks

Git pre-commit hook (`.githooks/pre-commit`) runs on staged `.go` files:
1. `gofmt -w` — auto-format
2. `golangci-lint` — linting
3. `go vet` — static analysis
4. `go build ./...` — compilation check

Install: `cp .githooks/pre-commit .git/hooks/pre-commit && chmod +x`

## Tech Stack

- Go 1.26+
- Standard library only
- AST parsing via `go/parser`
- Templates via `text/template`
- Concurrency via `sync.WaitGroup` (4 goroutines for parsing, 1 for scanning)

## Code Style

**Import aliases:** snake_case package names must use camelCase aliases.

```go
// RIGHT
import fileScanner "github.com/diplexhq/diplex/internal/file_scanner"
```

## Error Handling

**Panic is the only error mechanism:**

- Failures at generation time = fatal by nature
- Use `panic("message")` — flag errors, missing files, parse failures
- Never use `os.Exit()` directly
- Use `utils.Must()` and `utils.NoErr()` liberally

When encountering malformed Go code, always panic with a descriptive message:

```go
panic("corrupted Go source: empty identifier name — fix your source code")
```

## Architecture Overview

```
main.go
  └─ diplex.NewDI() → DI facade (config, logger, scanner, parser, resolver, generator)
       │
       ├─ Scanner.Scan() → domain.SourceFiles (channel)
       ├─ Parser.Parse(SourceFiles) → domain.ParsedData (providers + interfaces)
       ├─ Resolver.Resolve(ParsedData) → domain.ResolvedData
       └─ Generator.Generate(ResolvedData) → .go files
```

### Parser
Analyzes Go source files and extracts: providers (`New*` functions), interfaces, implementations (struct methods with `RealType: true` flag), type aliases, and embeds.

### Resolver
Builds indexes (`typeIndex`, `interfaceMethodIndex`) and resolves all provider dependencies for DI facades via BFS traversal.

### Generator
Produces DI container code from resolved data via Go templates (`facade.tmpl`, `provider.tmpl`, `head.tmpl`, `import.tmpl`).

## Generated Code Hash Verification

The `internal/tests/di_test.go` integration test generates a DI container and verifies its SHA-256 hash.

- Hash is stored as `expectedHash` constant
- If generator logic changes, the test fails with a hash mismatch message
- To update: regenerate via `go run .` → `sha256sum internal/tests/generated/diplex/di.go` → update `expectedHash` constant

This ensures deterministic generation and prevents accidental changes to the generator code.

## Self-Validation

```bash
# Build DI for the project itself (no extra flags, uses internal/di and internal/generated)
go run .

# Build DI for test matrix (specify target directory via -scan)
go run . -scan internal/tests -di internal/tests/di -out internal/tests/generated/diplex
```

Both must compile successfully (`go build ./...`).
