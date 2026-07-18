# Project Architecture

## Pipeline Overview

```
go/parser → AST → Scanner → chan SourceFile → Parser → ParsedData → Resolver → ResolvedData → Generator → di.go
```

The pipeline is a four-stage flow: **scan → parse → resolve → generate**. Each stage produces a well-defined intermediate representation consumed by the next.

## Stage 1: Scanner (`internal/scanner/`)

Single goroutine walks directories via `filepath.WalkDir`, applies skip pattern filters, and sends accepted `.go` file paths to a buffered channel (`chan SourceFile`, capacity 4). Parallelism is deferred to the parser stage.

## Stage 2: Parser (`internal/parser/`)

Four goroutines consume the channel concurrently, parsing Go AST and extracting providers (`New*` functions), interfaces, implementations (receiver methods), type aliases, and embeds. Mutable state is protected by `sync.Mutex`. Post-parse, aliases are flattened and embeds are BFS-flattened.

## Stage 3: Resolver (`internal/resolver/`)

Builds two indexes (`typeIndex` + `interfaceMethodIndex`) from parsed data and resolves dependencies via BFS traversal starting from facade method result types. Supports generic constraint narrowing and combinatorial resolution.

> **Deep internals** — algorithms, constraint narrowing, combinatorics, performance characteristics: see [ARCHITECTURE.md](ARCHITECTURE.md).

## Stage 4: Generator (`internal/generator/`)

Produces DI container `.go` files from `ResolvedData` using four embedded templates (`head`, `import`, `facade`, `provider`). Provider fields are wrapped in `sync.OnceValue()` for lazy initialization. Output is deterministic (SHA-256 verified).

## Design Principles

- **Interface-based DI**: Dependencies are wired through interfaces defined in `-di` directories (facades). The generated code implements these facades.
- **Narrow interfaces at point of use**: Each consumer defines the minimal interface it needs. The resolver matches implementations via method signature comparison.
- **Name-based disambiguation**: When multiple providers satisfy a type, parameter names resolve ambiguity via `ResultName` matching.
- **Slice aggregation**: `[]T` parameters automatically collect all matching providers.
- **Zero external dependencies**: Standard library only (`go/parser`, `go/types`, `text/template`, `sync`).

## Type Alias Resolution Pipeline

```
parseTypeAliases → resolveAliases → resolveProviders → resolveMethods
                    │                │                  │
                    │                │                  └─ replace Params & Results
                    │                └─ replace Arguments & Result
                    └─ BFS-flatten alias chains
```

Built-in aliases (`byte`→`uint8`, `rune`→`int32`, `interface{}`→`any`) are resolved in `parseTypeAliases`, then applied transitively to all provider arguments, results, and method signatures during `resolveAliases`.

## Method Signature Matching

The resolver matches interface methods using two-tier key lookup in `interfaceMethodIndex`:
1. **Full signature key** — `"Name(args)(results)"` for non-generic methods
2. **Bare name key** — `"Name"` for generic methods

Both are queried and results unioned. This ensures both concrete implementations and generic providers satisfy interface requirements.
