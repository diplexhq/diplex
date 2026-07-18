# Developer Guide

## Principles

- **Zero dependencies** — standard library only. No external packages.
- **Test-driven development** — every feature must be implemented and covered in `internal/tests/`. Each scenario is a source file with a provider, a narrow interface at point of use, and a corresponding DI wiring in the generated `di.go`.
- **Self-hosting** — DIplex generates DI container for itself. Both `main.go` and `internal/tests/` use the same generator.
- **Minimalism** — no over-engineering. One concern per file.
- **Type-safety** — everything checkable at compile-time.
- **Readability** — generated code must be self-documenting.

## CI / Git Hooks

Git pre-commit hook (`.githooks/pre-commit`) runs on staged `.go` files:
1. `golangci-lint` — linting (auto-format via `gofumpt`)
2. `go vet` — static analysis
3. `go build ./...` — compilation check

Install: `cp .githooks/pre-commit .git/hooks/pre-commit && chmod +x`

## Tech Stack

- Go 1.26+
- Standard library only: `go/parser`, `go/types`, `text/template`, `sync`, `context`, `os`, `io`
- No reflection, no code generation outside diplex itself

## Code Style

**Import aliases:** snake_case package names must use camelCase aliases.

```go
// CORRECT
import astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
```

**Error handling:** panics are the primary error mechanism. Use `panic("message")` for generation failures. Use `utils.Must()` and `utils.NoErr()` liberally. `os.Exit()` is only in `main.go` defer recovery.

```go
panic("corrupted Go source: empty identifier name — fix your source code")
```

## Architecture Overview

```
main.go
  └─ diplex.NewDI() → DI facade (config, logger, scanner, parser, resolver, generator) — generated in `internal/generated/diplex/di.go`
        │
        ├─ Scanner.Scan() → domain.SourceFiles (chan SourceFile)
        ├─ Parser.Parse(SourceFiles) → domain.ParsedData (providers + interfaces)
        ├─ Resolver.Resolve(ParsedData) → domain.ResolvedData
        └─ Generator.Generate(ResolvedData) → .go files
```

Each stage is a separate package (`internal/scanner/`, `internal/parser/`, `internal/resolver/`, `internal/generator/`). Deep internals of each stage — algorithms, data structures, and complexity — are documented in [ARCHITECTURE.md](ARCHITECTURE.md).

### Key design decisions

- **Scanner** produces a buffered channel (`chan SourceFile`, capacity 4). Files are walked by a single goroutine; parallelism is deferred to parsing.
- **Parser** uses exactly 4 worker goroutines sharing a mutex-protected `parseState`. Post-parse, aliases and embeds are resolved in sequence.
- **Resolver** builds a two-index system (`typeIndex` + `interfaceMethodIndex`) and resolves dependencies via BFS from facade methods. Constraint narrowing and combinatorial generic resolution are documented in [ARCHITECTURE.md](ARCHITECTURE.md).
- **Generator** renders four templates (`head`, `import`, `facade`, `provider`) into deterministic output verified by SHA-256 hash.

## Testing

All DI scenarios must be implemented and covered in `internal/tests/`. The `internal/tests/di_test.go` integration test verifies deterministic generation via SHA-256 hash of `di.go`.

To update hash after generator changes:
```bash
go run . -scan internal/tests -di internal/tests/di -out internal/tests/generated/diplex
sha256sum internal/tests/generated/diplex/di.go
# update expectedHash in internal/tests/di_test.go
```

## Self-Validation

```bash
# Build DI for the project itself
go run .

# Build DI for test matrix
go run . -scan internal/tests -di internal/tests/di -out internal/tests/generated/diplex
```

Both must compile: `go build ./...` must pass.
