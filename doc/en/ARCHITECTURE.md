# Architecture Deep Dive

This document describes the internal architecture of DIplex in detail. It covers scanning, concurrency, parsing, the two-index resolver architecture, constraint narrowing, and generic combinatorics.

## Pipeline Stages

```
go/parser → AST → Scanner → chan SourceFile → Parser → ParsedData → Resolver → ResolvedData → Generator → di.go
```

The pipeline is a four-stage flow: **scan → parse → resolve → generate**. Each stage produces a well-defined intermediate representation consumed by the next.

## Stage 1: File Scanning

### Entry Point

`Scanner.Scan()` (`internal/scanner/scan.go`) returns a `chan SourceFile` (type alias for `string`, representing an absolute file path). The channel has buffer capacity of 4.

### Walking Logic

A single goroutine iterates over configured `scanDirs`. For each directory, `filepath.WalkDir()` recursively descends the tree. At every entry, three filters are applied:

1. **Directory skip** — if `skipPattern.MatchString(path)` on a directory, `filepath.SkipDir` is returned, preventing recursion into that subtree.
2. **Suffix filter** — non-`.go` files are skipped entirely.
3. **Path filter** — if `skipPattern.MatchString(path)` on a file, the file is skipped.

```
filepath.WalkDir(dir, func(path, DirEntry, error) {
    if d.IsDir() && skipPattern.Match(path) → filepath.SkipDir
    if !strings.HasSuffix(path, ".go") → continue
    if skipPattern.Match(path) → continue
    ch <- SourceFile(path)
})
```

The default skip pattern is `` `(internal\/generated\/diplex|tests|mocks?|_test\.go|_mock\.go)$` ``. It can be overridden via the `-skip` flag.

### Concurrency Model

The scanner itself is **single-threaded** — one goroutine walks all directories and sends files to the channel. Parallelism is deferred to the parser stage via 4 concurrent consumer goroutines. This avoids per-file goroutine overhead while keeping the channel buffer small (capacity 4).

### Edge Cases

- `internal/di` is not in the default scan directory and must be explicitly added via `-scan` flag.
- Generated files are always excluded by the default skip pattern.
- The channel is closed after all directories are fully walked, signaling EOF to consumers.

## Stage 2: AST Parsing

### Concurrent Parsing

The parser spawns exactly **4 worker goroutines** (`parser.go:65`). Each worker ranges over the `SourceFiles` channel until it is closed, calling `parseFile()` for each source.

### Mutable State

`parseState` (`parser.go:32-40`) holds all intermediate data:

```go
type parseState struct {
    providers     domain.Providers          // ProviderID → *Provider
    interfaces    domain.Interfaces         // InterfaceID → InterfaceInfo
    embeds        map[InterfaceID][]InterfaceID  // interface → embedded interfaces
    typeAliases   map[string]string         // pkg.Alias → pkg.BaseType
    resolvedTypes map[string]string         // cache for alias resolution
    mu            sync.Mutex                // protects concurrent access
    fSet          *token.FileSet            // shared AST file set
}
```

All mutations to `providers`, `interfaces`, `embeds`, `typeAliases`, and `resolvedTypes` are serialized through `state.mu.Lock()`.

### Entity Extraction

#### Providers (`parse_provider.go`)

A function qualifies as a DI provider when ALL conditions are met:

| Condition | Check |
|-----------|-------|
| No receiver | `funcDecl.Recv == nil` |
| Name prefix | `strings.HasPrefix(name, "New")` |
| Result count | `len(results) == 1 or 2` |
| First result exported | `utils.IsExported(baseIdent)` |
| Optional error | If 2 results, second must be `error` |

**Generic parameter extraction** (`extractGenericParams`):
- Single type param → alias `T`, suffix `[T]`
- Multiple type params → aliases `T`, `T1`, `T2`, ... suffix `[T, T1, ...]`
- Union constraints (`User | Order`) → flattened via `walkConstraint()` into `[]string{"User", "Order"}`
- Interface constraints (e.g., `fmt.Stringer`) → stored as single-element list

#### Interfaces (`parse_interface.go`)

For each `*ast.InterfaceType`:
- Methods are serialized via `astStringer.FieldsToStrings()` — **parameter names are discarded**, only types matter
- Embedded interfaces recorded in `state.embeds` map
- `RealType: false` marks declared interfaces vs. implementations

#### Receiver Methods (`parse_receiver.go`)

Methods on concrete types become interface implementations:
- Exported methods on exported structs
- Receiver can be pointer or value type
- Generic receivers produce `[T]` suffixes and type aliases
- `RealType: true` distinguishes concrete types from interface declarations

#### Type Aliases (`parse_type_aliases.go` + `resolve_aliases.go`)

- `type X = Y` declarations populate `state.typeAliases`
- Built-in aliases pre-loaded: `byte→uint8`, `rune→int32`, `interface{}→any`
- `resolveAliases()` flattens alias chains transitively (A→B→C becomes A→C) and detects cycles

### Post-Parse Resolution

After all files are parsed (all 4 workers done):

1. **Alias resolution** (`resolveAliases`) — BFS-flatten alias chains, replace all types in providers, methods, and interface contracts
2. **Embed resolution** (`resolveEmbeds`) — BFS-flatten embedded interfaces, copying methods from embeds to parent. Missing embedded interfaces cause the parent to remain but with fewer methods.

### AST Canonicalization

The `ast_stringer/` package (private) converts AST nodes to canonical Go syntax strings for comparison. Key behaviors:

- **Identifier resolution order**: import alias → generic alias → builtin type → package prefix
- **Interface method sorting**: methods are sorted alphabetically for canonical representation (Go interface equality ignores method order)
- **Composite types**: `chan<- Event`, `map[string][]int`, `func(string) error` — all serialized to deterministic string form

## Stage 3: Dependency Resolution

### Two-Index Architecture

The resolver builds two complementary indexes from `ParsedData` (`build_index.go`):

#### Type Index (`byType`)

```go
map[string][]*domain.Provider  // normalized result type → candidate providers
```

Builds a key by calling `normalizeGenericParameter()` on each provider's `Result`:
- `RedisStorage[User, Order]` → `RedisStorage[T, T]`
- `RedisStorage[T, T1]` → `RedisStorage[T, T]`
- `Cache[string, int]` → `Cache[T, T]`

Every provider is indexed under its normalized key. Non-generic providers are also indexed by their exact result.

#### Interface Method Index (`byMethod`)

```go
map[string][]*domain.Provider  // method signature → providers implementing it
```

For each provider's result type, if that type implements an interface, each method is indexed by a `methodKey()`:

| Method type | Index key example |
|-------------|-------------------|
| Non-generic | `Get(id int) (*User, error)` |
| Generic | `Get` (bare name only) |

Both full and bare keys are queried during resolution.

### Facade Resolution

An interface qualifies as a DI facade when its `InterfaceID` starts with `module/DIDir/` (`resolve_facades.go`). For example, with module `github.com/diplexhq/diplex` and `-di internal/di`, any interface at `github.com/diplexhq/diplex/internal/di/...` becomes a facade. Each facade method's result type seeds the BFS queue.

### BFS Provider Resolution

The provider resolution uses a stack-based BFS (`resolve_providers.go`):

1. Extract all unique facade method result types into `queue`
2. Pop from stack, call `findProviderCollection()` for each
3. For each found provider, add its argument types to the queue (if not seen)
4. Repeat until queue is empty

```go
for len(queue) > 0 {
    arg := queue[len(queue)-1]
    queue = queue[:len(queue)-1]
    collection := findProviderCollection(arg, parsedData, index)
    providers[arg] = collection
    for _, p := range collection.Providers {
        for _, a := range p.Arguments {
            if !seen[a] { seen[a] = true; queue = append(queue, a) }
        }
    }
}
```

### Provider Collection Dispatch

`findProviderCollection()` (`find_providers.go:14-38`) handles slice vs. single resolution:

- If arg starts with `[]` → strip prefix, recurse, wrap result in `ProviderCollection{CollectionType: "slice", ...}`
- Nested collections panic
- Single providers are sorted by ID for deterministic output

Dispatch to type or interface matching via `findProviders()`:

```go
if parsedData.Interfaces[arg].RealType == false && has methods → findProvidersByInterface()
else → findProvidersByType()
```

### Generic Parameter Normalization

`normalizeGenericParameter()` (`utils.go:32-77`) replaces all generic type arguments inside `[...]` with `"T"`. It handles nested generics, pointers, and composite types:

```
pkg.Cache[order.Order, payment.Payment]  → pkg.Cache[T, T]
*pkg.Cache[T0, T1]                       → *pkg.Cache[T, T]
pkg.Cache[repo.Repo[T0], T1]             → pkg.Cache[T, T]
map[int]string                           → map[int]string  (map brackets are skipped)
```

### Type Matching Algorithm

`findProvidersByType()` (`find_providers_by_type.go`):

1. Normalize the wanted type using `normalizeGenericParameter()`
2. Look up in `typeIndex`
3. For each candidate, clone its generic constraints
4. Run `compareParams()` to match provider result against wanted type
5. If match, call `resolveProvider()` to generate concrete instantiations

#### compareParams — Character-by-Character Token Matching

`compareParams()` (`find_providers.go:60-90`) checks if a provider's result type satisfies a wanted parameter type:

1. **Exact match** — if strings are identical, return `true`
2. **Empty constraints** — if no generic constraints available and strings differ, return `false`
3. **Token loop** — consume both strings character-by-character:
   - Get next token from both strings
   - If provider token is **non-generic** (not `T`, `T0`, etc.) → must exactly match wanted token
   - If provider token is **generic** (`T`/`T0`/`T1`) → extract the concrete type from wanted string and `squeezeConstraint()`
4. **Both consumed** — after the loop, both strings must be fully consumed

#### squeezeConstraint — Generic Constraint Narrowing

`squeezeConstraint()` (`find_providers.go:95-107`) is the core of generic resolution:

- If the parameter name already exists in constraints → verify the concrete type is in the allowed list (or is `"any"`), then **narrow** the list to contain only that concrete type
- If the parameter name is new → create entry with the concrete type

This is a **progressive narrowing** process. Each `compareParams()` call along the BFS path further restricts the constraint map. By the time all methods are matched, constraints contain the concrete types needed for instantiation.

```go
func squeezeConstraint(constraints, paramName, concreteType) bool {
    if existing, ok := constraints[paramName]; ok {
        if !Contains(existing, concreteType) && !Contains(existing, "any") {
            return false  // constraint violation
        }
        constraints[paramName] = []string{concreteType}  // narrow to single
    } else {
        constraints[paramName] = []string{concreteType}  // new entry
    }
    return true
}
```

### Interface Matching Algorithm

`findProvidersByInterface()` (`find_providers_by_interface.go`):

1. Look up both full method key and bare name key in `byMethod` index
2. For each candidate:
   - If candidate result type equals wanted → direct match
   - If candidate is generic → extract prototype, clone constraints, narrow via `methodMatches()`
3. `methodMatches()` iterates all wanted methods, verifying:
   - Provider has each method
   - Argument count matches
   - Result count matches
   - Each argument/result type passes `compareParams()` with current constraints

#### Constraint Propagation for Interfaces

When matching generic providers against interfaces:

```go
// Step 1: Clone provider's original constraints (from type param declaration)
providerConstraints := maps.Clone(candidate.Generic)

// Step 2: Narrow via method matching results
for k, v := range constraints {  // constraints from instantiate generic
    if originalParam, ok := providerConstraints[v[0]]; ok {
        providerConstraints[originalParam] = narrowed[k]
    }
}

// Step 3: Generate all combinations and create concrete providers
resolveProvider(candidate, providerConstraints)
```

### Combinatorial Generic Resolution

When a generic provider's constraints contain multiple options (e.g., `K: ["string", "int"]`, `V: ["User", "Order"]`), `generateCombinations()` (`find_providers.go:150-187`) produces the Cartesian product:

```
K ∈ {string, int}
V ∈ {User, Order}

Combinations:
  1. {K: string, V: User}
  2. {K: string, V: Order}
  3. {K: int, V: User}
  4. {K: int, V: Order}
```

The combination generator uses recursive DFS over **sorted** constraint keys for determinism:

```
generate(idx, current):
    if idx == len(keys): result.push(clone(current)); return
    k := keys[idx]
    for each option in constraints[k]:
        current[k] = option
        generate(idx+1, current)
```

Each combination produces a concrete `*domain.Provider` via `utils.ReplaceTokens()`, substituting generic names with concrete types in Result, Arguments, Name, and ID.

### Name Disambiguation

When multiple providers match a single parameter type (`len(collection.Providers) > 1`), `selectSingleProvider()` (`resolve.go:95-117`) resolves ambiguity:

1. If collection has exactly 1 provider → return it
2. If `argName` is empty → return nil (caller panics)
3. Match `provider.ResultName` against `argName` using `matchByName()`:
   - Must have equal length
   - First character case-insensitive match
   - Rest of characters exact match
4. If exactly 1 match → return it; otherwise → nil (caller panics)

`ResultName` derivation: first named return value, or local name of result type (`*UserRepo` → `UserRepo` → `userrepo` lowercase).

## Stage 4: Code Generation

### Template System

Four embedded templates are combined per DI facade directory (`internal/generator/tmpl/`):

| Template | Purpose |
|----------|---------|
| `head.tmpl` | Package declaration and file comment |
| `import.tmpl` | Grouped imports with unique aliases |
| `facade.tmpl` | Facade method implementations delegating to DI struct |
| `provider.tmpl` | DI struct with `func()` fields + `sync.OnceValue` wrappers |

### Dependency Data Construction

`buildData()` performs a **BFS traversal** of the dependency graph starting from facade methods:

1. For each facade method, identify the resolved provider
2. Recursively traverse `ArgumentProviders` for each provider
3. Track unique packages, providers, and methods
4. Sort all collections deterministically (providers by ID, methods alphabetically)

### Package Alias Resolution

Imports are grouped and sorted:

1. **Standard library** — no alias needed if base name matches package path suffix
2. **Third-party** — alphabetical, with `idUniq()` preventing naming collisions
3. **Local packages** — after module prefix stripped, same dedup logic

### Provider Field Generation

Each provider field in the DI struct is a `func() ResultType` wrapped in `sync.OnceValue()`:

```go
di.entityNewUserRepo = sync.OnceValue(func() *entity.UserRepo {
    return entity.NewUserRepo(
        di.storageNewRedisStorageEntityUserString(),
        di.storageNewStringEntityUser(),
    )
})
```

- Providers returning `error` include: `result, err := fn(); if err != nil { panic(err.Error()) }; return result`
- Slice providers return `[]T{di.provider1(), di.provider2(), ...}`
- Field names are sanitized to camelCase: `entity.NewRepo[entity.Order]` → `entityNewRepoEntityOrder`

### Determinism Guarantees

- All map iterations are replaced with sorted key traversal
- Generic constraint keys are sorted before combinatorial generation
- Provider collections are sorted by ID before template rendering
- Import aliases use `idUniq()` for collision-free naming
- Result: identical source → identical `di.go` output (verified by SHA-256 hash)

## Data Flow Summary

```
Scanner.Scan()
  │  single goroutine, filepath.WalkDir, skipPattern filters
  ▼
chan SourceFile  (buffered, capacity 4)
  │
  ├─ 4 parser goroutines (sync.WaitGroup)
  ├─ parseState protected by sync.Mutex
  ├─ Post-parse: resolveAliases() + resolveEmbeds()
  ▼
ParsedData { Providers, Interfaces }
  │
  ├─ buildTypeIndex() → normalizeGenericParameter()
  ├─ buildInterfaceIndex() → methodKey()
  ├─ resolveFacades() → DIDir prefix match
  ├─ BFS queue: facade results → provider args → ...
  ├─ compareParams() → squeezeConstraint() → progressive narrowing
  ├─ generateCombinations() → Cartesian product of constraints
  ├─ selectSingleProvider() → matchByName() disambiguation
  ▼
ResolvedData { ResolvedFacades, ResolvedProviders }
  │
  ├─ buildData() → BFS dep tree, sorted imports, sorted providers
  ├─ 4 templates: head + import + facade + provider
  ├─ sync.OnceValue() lazy init wrappers
  └─ SHA-256 deterministic verification
  ▼
di.go (output)
```

## Performance Characteristics

| Stage | Complexity | Notes |
|-------|-----------|-------|
| Scanner | O(n) where n = files in tree | Single goroutine, buffered channel |
| Parser | O(m × f) where m = files, f = AST nodes | 4 workers, mutex-protected state |
| Index building | O(p × k) where p = providers, k = methods | Linear scan of parsed data |
| BFS resolution | O(d × c) where d = dependency depth, c = candidates | Short-circuits on exact match |
| Constraint narrowing | O(t × n) where t = type tokens, n = constraints | Character-by-character scan |
| Combinatorial gen | O(∏cᵢ) where cᵢ = constraint cardinality | Bounded by constraint size |
| Generation | O(p × t) where p = providers, t = templates | Sorted iterations, deterministic |

## Error Boundaries

| Scenario | Error mechanism | Recovery |
|----------|----------------|----------|
| No provider found | `panic("cannot generate DI: no providers for ...")` | None — build fails |
| Ambiguous match | `panic("no single provider found for ...")` | None — rename parameter |
| Cycle detection | None (runtime panic via sync.OnceValue) | Redesign dependency graph |
| Invalid AST | `astStringer` returns empty string | Skip entity |
| Alias chain cycle | `utils.ResolveReplacements()` panics | Fix alias declaration |
