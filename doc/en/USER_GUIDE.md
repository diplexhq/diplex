# User Guide

## Installation

```bash
go get -tool github.com/diplexhq/diplex@latest
go mod vendor
```

## Quick Start

```bash
go tool diplex
```

## CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-scan` | Comma-separated directories to scan for providers, interfaces, and implementations | `"internal"` |
| `-out` | Output directory for generated DI container files | `"internal/generated/diplex"` |
| `-di` | Comma-separated directories containing DI facade interfaces (the `DI` interface your app depends on) | `"internal/di"` |
| `-module` | Explicit Go module path (overrides automatic detection from `go.mod`) | _(auto from `go.mod`)_ |
| `-skip` | Regular expression for files/directories to skip during scanning | `"(internal\/generated\/diplex|tests|mocks?|_test\.go|_mock\.go)$"` |
| `-v` | Verbose output — prints generation details to stderr | — |
| `-s` | Silent mode — suppresses all output | — |

**Mutual exclusivity:** `-v` and `-s` cannot be used together — the tool will panic if both are set.

### Examples

```bash
# Scan specific directories
go tool diplex -scan internal/pkg,internal/service

# Custom output directory
go tool diplex -out pkg/generated/di

# Skip generated code directories
go tool diplex -skip "(generated|vendor|_test\.go)$"

# Override module path (no go.mod)
go tool diplex -module example.com/my/app

# Verbose mode
go tool diplex -v

# Silent mode
go tool diplex -s
```

## Workflow Order

DIplex performs work in four phases:

### 1. Scanning

Recursively scans the `internal` folder, ignoring mocks and test files.

### 2. Recognition

DIplex identifies two categories of entities:

**Providers** — functions matching the `New*` pattern that return an exported type. The first return value is the produced object; the optional second return value is `error`. Generic providers are supported.

```go
// Recognized as provider
func NewConfig() *Config
func NewDB(cfg *Config) (*DB, error)
func NewRepo[T User | Order]() *Repo[T]  // → NewRepo[User], NewRepo[Order]
```

**Interfaces** — both pure declarations and concrete type implementations. Embedding is supported for both interfaces and structs; generics are supported for type implementations.

```go
// Recognized as interface
type UserStore interface {
	Get(id int) (*User, error)
}
```

```go
// Recognized as implementation of an interface
type userRepository struct { db *DB }
func (r *userRepository) Get(id int) (*User, error) { ... }
```

### 3. Resolution

Dependencies are resolved automatically. The resolver matches providers to consumer parameters by type and parameter name.

- **Cyclic dependencies** are not detected — if a cycle exists, initialization panics at runtime.
- **Slice dependencies** collect all matching providers automatically.
- **Multiple implementations** — if a dependency can be satisfied by multiple providers, ambiguity is resolved by matching the provider's result name with the consumer's parameter name. If ambiguous or unmatched — panic.

```go
// Ambiguity resolution by parameter name
func NewUserRepo(userRepo *UserRepo) { ... }
// resolver matches `userRepo` parameter to `UserRepo` result

func NewStorage(redisDsn Dsn) { ... }
// resolver matches `redisDsn` parameter to `NewRedisDsn` which returns `(redisDsn Dsn)`
```

### 4. Generation

Generates stable code in `-out`, creating implementations for facade interfaces.

## Result

After successful execution, `.go` files appear in `-out`


- Structure with fields for each resolved dependency
- `New*()` constructors creating facade implementations
- `sync.OnceValue`-based lazy initialization — dependencies are created only on first access

One DI facade can be reused in different builds (web server, cron, CLI, queue processor). 
For convenience, you can define separate facades per build, each with only the required providers.

### Using the Generated Container

```go
// In your project — define the facade
// internal/di/di.go
package di

type DI interface {
    Config() *config.Config
    DB() *db.DB
    UserRepository() *userRepository
}
```

```go
// After running `go tool diplex`:

func main() {
	deps := diplex.NewDI()  // all dependencies already wired

    cfg := deps.Config()
    db := deps.DB()
    repo := deps.UserRepository()

    // ... use services
}
```

## Limitations

### Type Matching

- **Primitive types are matched by name** — `string`, `int`, `int64`, `bool` and other built-in primitives are matched by argument name and result name/type. Use a named type or alias instead of a bare type (e.g., `type Port int` or `type Port = int`), not raw `int` or `string`.
- **External interfaces are not scanned** — only interfaces declared in project source files are considered. Standard library and third-party package interfaces are invisible to the resolver.

### Provider Rules

- **Return signature** — providers must return exactly 1 or 2 values. If 2, the second must be `error`. Functions returning 3+ values are silently ignored.
- **Singletons only** — all providers are instantiated once and cached via `sync.OnceValue`. No factory / per-call semantics.
- **Cyclic dependencies** — cause panic at runtime due to infinite constructor recursion. The resolver does not detect cycles at compile time; this is a known limitation of the BFS-based resolution approach.

### Interface Rules

- **Full narrow interface support** — narrow interfaces at point of use are supported, but for overly narrow ones you must ensure disambiguation by argument name is possible (e.g., different parameter names for different implementations of the same interface).
- **Tags, weights, priorities via interface** — DIplex has no built-in tags, weights, or priorities. If you need such features, define an interface with methods `ID() string`, `Name() string`, `Weight() int`, `Priority() int` etc., implement it on your types, and in a provider request a slice (`[]MyInterface`). The resolver will collect all matching providers into a list you can sort or filter at runtime.

### Rules for generic types

- **Ensure the ability to narrow to a concrete variant** — the combinatorial list of generic providers is supported, but must be reduced to a finite number of combinations. This is a rather strange thing and requires real-world scenarios to work through.

### Composite types

- **Complex inline types** — the resolver matches `map[string]string`, `chan<- Event`, `<-chan Result` and other composite types in method signatures. Complex inline types with functions, interfaces, or structs are also matched. However, it is recommended to avoid such complexity — this is poor style, rarely used, and therefore not covered by tests. Named types are matched automatically; type aliases do not change resolver behavior.

## Full scenario coverage

All supported DI patterns are documented in [SCENARIOS.md](SCENARIOS.md).
