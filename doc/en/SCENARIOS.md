# DI Scenarios

Complete catalog of all supported DI patterns.

## Provider Patterns

| # | Pattern | Example |
|---|---------|---------|
| 1 | Root provider (no deps) | `func NewConfig() *Config` |
| 2 | Single-level dependency | `func NewDBConfig(cfg *Config) DBConfig` |
| 3 | Struct method as provider | `func (c *Config) DB() DBConfig` |
| 4 | Generic provider | `func NewRepo[T any]() *Repo[T]` |
| 5 | Generic constraint matching | `func NewRepo[T User \| Order]() *Repo[T]` |
| 6 | Generic 2 type params | `func NewCache[V, K]() *Cache[V, K]` |
| 7 | Provider with error | `func NewHealthChecker(...) (*HealthChecker, error)` |
| 8 | Named primitive types | `type Port int; func NewPort() Port` |
| 9 | Value type provider | `func NewDBConfig(...) DBConfig` |
| 10 | Factory pattern | `type FactoryProvider func(...) Handler` |
| 11 | Multiple concrete impls | First alphabetically by ID |

## Interface & Type Patterns

| # | Pattern | Example |
|---|---------|---------|
| 12 | Narrow interface per consumer | `type Repo interface { Get(...) }` |
| 13 | Slice collection | `[]handler.Handler` aggregated |
| 14 | Struct embedding | `type Handler struct { Base }` |
| 15 | Interface composition via embedding | `type OrderService interface { OrderReader; OrderWriter }` |
| 16 | Import alias (package) | `import testLogger "github.com/..."` |
| 17 | Aliased narrow interface | `type User = entity.User` |
| 18 | Nested alias chain | `type Order = entity.OrderAlias` |
| 19 | Mixed aliased generics | `type UserStatsRepo = Repo[User]` |
| 20 | Import alias interface embedding | `import api "..."` + `type Manager interface { api.Store }` |
| 21 | Pointer embed in struct | `type Tracker struct { *domain.Timestamps }` |
| 22 | Interface subset matching | `Repo.Get` ⊂ `Repo[T].Get` |

## Advanced Type Patterns

| # | Pattern | Example |
|---|---------|---------|
| 23 | Chan direction types | `chan<- Item`, `<-chan []Item` |
| 24 | Map types in methods | `ListItems(filter map[string]string)` |
| 25 | Variadic methods | `Sum(items ...*Item) error` |
| 26 | Named return values | `Report() (count int, total float64, err error)` |
| 27 | Cross-service dependency chain | `Service` → `Repo` + `Gateway` + `Validator` |
| 28 | Constructor with 4+ deps | `NewService(a, b, c, d)` |
| 29 | Function type dependency | `callback.HandlerFunc` |
| 30 | Channel type dependency | `event.NotifyChan` |
| 31 | Any/interface{} normalization | `type FilterArgs any` |
| 32 | Cross-module dependency chain | `OrderService` → `PaymentService` → `StripeGateway` |
| 33 | Nested generic instantiation | `Repository[T]` → `InMemoryStore[Order]` → `Service` |

## Resolution & Filtering

| # | Pattern | Description |
|---|---------|-------------|
| 34 | Private type filtering | Private structs excluded from scanning |
| 35 | DI facade interface | `di.DI` interface with methods |
| 36 | Handler slice aggregation | All handlers collected into `[]handler.Handler` |

## Limitations

### Unresolved Issues

- **Narrow interface collision**: The DI resolver cannot distinguish narrow interfaces of different consumers with identical method signatures. If two handlers require `Repo.Get(id int) (entity.Order, error)` and `Repo.Get(id int) (entity.User, error)` — the conflict is unresolved. In such cases, wide interfaces (like `entity.Repository`) must be used.
- **Complex inline types**: `map[string]string`, `chan<- Event`, `<-chan ComplexResult` and other complex composite types in method signatures are not the resolver's goal. Named types and aliases (e.g., `type FilterArgs = map[string]string`) should be defined in separate packages.


### Provider Signature Rules

| Rule | Behavior |
|------|----------|
| Max 2 return values | Constructors returning 3+ values are **silently ignored** |
| Optional error | Second return value must be `error` type |
| Pointer or value | First return must be `*T` or `T` |

### Generic Type Rules

| Rule | Behavior |
|------|----------|
| Type param normalization | `T0`, `T1` → aligned to `T`, `T0`, `T1`... |
| Constraint widening | `[T User \| Order]` → concrete instantiations for each constraint |
| Empty constraints | Generic with no constraints produces **no** instantiations |

## Integration Tests

All scenarios covered by tests in `internal/tests/`:

```
internal/tests/
├── cache/         — Generic types, nested instantiations
├── config/        — Root providers, named primitives
├── di/            — DI facade interface
├── entity/        — Generic providers, narrow interfaces
├── handler/       — Slice aggregation, struct embed
├── logger/        — Private type filtering
└── metrics/       — 4+ dependency constructor
```

**Source of truth:** `internal/tests/di_test.go`.