# Changelog

## [2.0.1] - 2026-07-19

### Doc Fixes
- small updates

## [2.0.0] - 2026-07-18

### Breaking Changes

- **No ambiguity of multiple providers for the same dependency** — when multiple providers return the same type, the resolver now **requires** unambiguous name resolution via `ResultName` matching against consumer parameter names. `ResultName` is the first named return value, or the result type's local name if unnamed. If names match uniquely, the provider is selected. No fallback — resolution panics if no unique match is found.
- **Parameter name must not collide with result type name** — a provider's parameter name that matches the result name of another provider creates ambiguity. Use distinct names: e.g., `func NewUserRepo(db *dbDsn) *UserRepo` where `dbDsn` is the result name of the DB provider.

### Features

- **Provider name matching** — when multiple providers return the same type, the resolver disambiguates by matching the provider's `ResultName` (first named return value, or result type's local name if unnamed, e.g. `UserRepo` from `func NewUserRepo() *UserRepo`) against the consumer's parameter **name**. For facade methods, matches against the method name (lowercased first char). Names must have equal length, first character case-insensitive, rest exact. Panics if no unique match is found — ambiguity is architecturally unacceptable.
- **ResolvedData restructuring** — `ResolvedData` now uses `map[InterfaceID]ResolvedFacade` and `map[ProviderID]ResolvedProvider` instead of `map[Parameter]ProviderCollection`. Each resolved facade method and provider stores the exact resolved `*Provider` reference, eliminating redundant lookups during code generation.
- **Provider ID as full string** — `Provider.ID` is now a full string (`pkg.Name`) instead of a method `Id()`. All references updated across parser, generator, and resolver.
- **BFS resolution with visited set** — provider dependency resolution now uses BFS with a `visited` set, ensuring each provider is resolved exactly once and its transitive dependencies are traversed in dependency order.
- **ArgNames tracking** — providers now store `ArgNames` (parameter names from the function signature) for use in name matching during resolution.

### Removed

- `internal/tests/cache/` — removed (was testing cache provider, now consolidated)
- `internal/tests/entity/` — removed old entity types (`Entity`, `Order`, `Repo`, `User`), replaced by `entity/entity.go` with composite types
- `internal/tests/event/` — removed (consolidated into callback/event patterns)
- `internal/tests/handler/*` — removed `cache`, `detail`, `find`, `user/find`, `user/lookup` handler files
- `internal/tests/entity/repo.go` — removed (consolidated into `entity/entity.go`)
## [1.0.1] - 2026-07-13

### Bug Fixes

- Support hyphens (`-`) in package paths (e.g. `my-package`) for correct alias resolution
- `NewDI()` now returns the facade interface type instead of struct pointer
- All logs redirected to stderr; only `Info("generation complete")` on stdout

## [1.0.0] - 2026-07-12

### Initial Release

- Zero-dependency Go DI container generator
- AST-based scanning and parsing
- Narrow interface resolution
- Generic type support
- Self-hosting (generates its own DI container)
- Integrated benchmarks