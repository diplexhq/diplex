# DIplex Requirements

## Functional Requirements

### F1: Code Scanning
- Recursive directory traversal
- Skip `_test.go` files
- Skip `mocks/` directories

### F2: Provider Recognition
- Functions with `New` prefix
- Returns pointer `*Type` or value `T`
- Optional `error` as second return type
- Generic providers support (`func NewRepo[T User | Order]() *Repo[T]`)
- Concrete instantiation from generic providers via constraint matching

### F3: Interface Recognition
- Parse interface methods
- Embedded interfaces from scanned source files (not stdlib/external)
- Generic interfaces with type parameters (`Repository[T any]`)

### F4: Implementation Recognition
- Public struct methods
- Match interfaces by signature (`compareArgs` with constraint narrowing)
- Pointer and value receiver support
- Generic receiver support (`*Repo[Order]`, `*InMemoryStore[T]`)
- Struct embedding

### F5: Code Generation
- DI facade generation (one `.go` file per `-di` directory)
- Dependency wiring via `sync.OnceValue`
- Slice collection generation (`[]Handler` → aggregated list)
- CamelCase field sanitization for generic provider IDs
- Unique package aliases with collision avoidance

### F6: Error Handling
- Panic on missing interface or implementation
- Panic-first policy (this is a codegen utility)

## Non-Functional Requirements

### NF1: Performance
- Parallel scanning (channel-based iterator)
- Concurrent file parsing (`sync.WaitGroup`)

### NF2: Reliability
- Integration tests via `internal/tests/` (source of truth)
- Self-validation: build DI for project + build DI for `internal/tests/`

### NF3: Usability
- CLI flags (`-scan`, `-out`, `-skip`, `-module`, `-di`, `-v`, `-s`)
- Verbose mode with generation details

## Architectural Constraints

### A1: Compatibility
- Go 1.26+ (current go.mod version)
- **No external dependencies** (stdlib only)
- GOOS/GOARCH support

### A2: Generated Code
- Must compile without errors
- Follow Go Code Review Comments
- gofmt formatting
- Compatible with go vet, staticcheck

### A3: Extensibility
- Templating for customization (`text/template`)
- DI facade pattern for extensibility
- Configuration via flags