# DIplex - Go DI Container Generator

## Purpose

Automatic Go code scanner and high-performance DI container generator at code generation time (code generation).

## Core Principles

- **Compile-time DI** - no runtime reflection, maximum performance
- **Type-safe** - all dependencies checked at compile time
- **Zero overhead** - generated code equals manual dependency injection
- **Convention over configuration** - uses standard Go patterns

## How It Works

### 1. Scanning
Recursive traversal of Go files in specified directories. Skips `_test.go` and `mocks/` directories. Parses AST via standard `go/parser`.

### 2. Analysis
Extracts three entity types:

**Providers:** Functions matching `New*` pattern, returning a pointer (optionally with `error`). Supports generics: `NewRepo[T User | Order]` produces concrete instantiations.

**Interfaces:** Declared interfaces with methods and embedded interfaces. Supports generics: `Repository[T any]`.

**Implementations:** Public struct methods, including generic receiver: `func (r *Repo[Order]) Get(id int) (Order, error)`.

### 3. Generation
Produces DI facade files in the output directory. Each `-di` directory produces one `.go` file with a `DI` struct and `NewDI()` constructor using `sync.OnceValue` for singleton semantics.

Field names for generic providers are sanitized to valid camelCase: `entity.NewRepo[entity.Order]` → `entityNewRepoEntityOrder`.

## Project Structure

```
├── main.go                    # Entry point, orchestration
└── internal/
    ├── generator/             # Code generation (templates: facade, provider, head, import)
    ├── resolver/              # Provider resolution and dependency mapping
    │   ├── build_index.go     # typeIndex + interfaceMethodIndex
    │   ├── find_providers.go  # Provider lookup and constraint matching
    │   └── resolve*.go        # BFS dependency traversal
    ├── scanner/               # Go file scanning
    ├── parser/                # AST parsing & analysis
    │   ├── ast_stringer/      # AST → string (canonical formatting)
    │   └── resolve*.go        # Aliases and embeds
    ├── domain/                # Domain types (ParsedData, Provider, MethodContract)
    ├── di/                    # DI facade interface
    ├── config/                # CLI flags and configuration
    └── utils/                 # Must, NoErr, SanitizeIdent, Logger
```

## Usage

```bash
# Install as Go tool
go get github.com/diplexhq/diplex

# Run from target project root (reads go.mod automatically)
go tool diplex

# Scan specific directories
go tool diplex -scan internal,pkg

# Custom output directory
go tool diplex -out internal/generated/diplex

# Explicit module path (skip go.mod)
go tool diplex -module example.com/my/project

# Custom skip pattern (regexp)
go tool diplex -skip "(internal\/generated\/diplex|testdata|mocks?|_test\.go|_mock\.go)$"

# Specify DI facade directories
go tool diplex -di internal/di

# Verbose / Silent mode
go tool diplex -v
go tool diplex -s
```

**Defaults:** scans `internal/`, DI facades in `internal/di/`, output to `internal/generated/diplex/`.

## Limitations

- **Primitive types must be wrapped** — `string`, `int` etc. cannot be matched. Each primitive argument must use a unique named type.
- **Narrow interfaces** — declare interfaces at point of use, not broad shared interfaces.
- **External interfaces** — only interfaces from project code are scanned, not from stdlib or external packages.
- **Provider returns 1-2 values** — only `T`, `T, error`, `*T`, `*T, error` supported.
- **All providers are singletons** via `sync.OnceValue`. Cyclic dependencies cause panic at runtime.

## Interface Style

```go
// GOOD — narrow consumer interface
type ScanDirProvider interface {
    ScanDirs() string
    Silent() bool
}

// BAD — broad interface
type GlobalConfig interface {
    ScanDirs() string
    OutputDir() string
    IsSilent() bool
}
```

Avoid `Get` and `Is` prefixes in method names. Use nouns or adjectives directly.

## Full Scenario Coverage

See [SCENARIOS.md](SCENARIOS.md) for the complete catalog of all supported DI scenarios and integration tests in `internal/tests/`.