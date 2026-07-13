# Changelog

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