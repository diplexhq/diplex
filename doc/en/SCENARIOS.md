# DI Scenarios

Full catalog of all DI patterns covered by `internal/tests/` and verified by the generated container in `internal/tests/generated/diplex/di.go`.

## 1. Root Providers (no dependencies)

`New*()` functions with zero parameters.

| Source | di.go field | Line | Type |
|--------|------------|------|------|
| `config.NewConfig() *Config` | `configNewConfig` | L74 | `*config.Config` |
| `logger.NewLogger() Logger` | `loggerNewLogger` | L208 | `logger.Logger` |
| `event.NewNotifyChan() NotifyChan` | `eventNewNotifyChan` | L112 | `chan Event` (channel) |
| `event.NewHandlerFunc() HandlerFunc` | `eventNewHandlerFunc` | L108 | `func(path string) string` (function type) |
| `event.NewComplexEvent() ComplexEvent` | `eventNewComplexEvent` | L104 | `interface` (chan direction types) |
| `metrics.NewPort() Port` | `metricsNewPort` | L224 | `int` (named primitive) |
| `metrics.NewTimeout() Timeout` | `metricsNewTimeout` | L228 | `time.Duration` (named primitive) |
| `storage.NewDbConnection(Dsn) *DbConnection` | `storageNewDbConnection` | L61, L232 | `*storage.DbConnection` (used by generics below) |

Files: `internal/tests/config/config.go:17`, `internal/tests/logger/logger.go:16`, `internal/tests/event/event.go:32`, `internal/tests/event/callback.go:6`, `internal/tests/metrics/metrics.go:13`, `internal/tests/metrics/metrics.go:19`, `internal/tests/storage/db_storage.go:11`

## 2. Single-level Dependency Providers

| Source | di.go field | Line | Dependency | Return type |
|--------|------------|------|------------|-------------|
| `config.NewDbDsn(*Config) (dbDsn Dsn)` | `configNewDbDsn` | L38, L78 | `configNewConfig` | `config.Dsn` (named `string`) |
| `config.NewRedisDsn(*Config) RedisDsn` | `configNewRedisDsn` | L39, L84 | `configNewConfig` | `config.RedisDsn` (alias `= Dsn`) |
| `storage.NewRedisConnection(redisDsn Dsn) *RedisConnection` | `storageNewRedisConnection` | L63, L244 | `configNewRedisDsn` | `*storage.RedisConnection` |
| `storage.NewDbConnection(dbDsn Dsn) *DbConnection` | `storageNewDbConnection` | L61, L232 | `configNewDbDsn` | `*storage.DbConnection` |

Files: `internal/tests/config/config.go:25,29`, `internal/tests/storage/redis_storage.go:9`, `internal/tests/storage/db_storage.go:11`

### DSN Resolution by Name

`metrics.NewHealthChecker(dbDsn, redisDsn config.Dsn, port Port)` — two `config.Dsn` params resolved by parameter name matching to provider result names:

- `dbDsn` → `config.NewDbDsn(...) (dbDsn Dsn)` — named return value `dbDsn` in `config/config.go:29`
- `redisDsn` → `storage.NewRedisConnection(redisDsn config.Dsn)` — named param in `storage/redis_storage.go:9`
- `config.RedisDsn = Dsn` — type alias in `config/config.go:7`, resolved as same type `config.Dsn`

## 3. Generic Providers: Chain with Constraint Widening

### 3a. `storage.NewRedisStorage[V any, K string \| int](conn *RedisConnection) *RedisStorage[K, V]`

Signature params: `[V any, K string | int]` — reversed order in instantiation: `NewRedisStorage[User, string]` at signature `[V, K]`.

| di.go field | Instantiation | K | V | Line | Consumer |
|------------|--------------|---|---|------|----------|
| `storageNewRedisStorageEntityUserString` | `RedisStorage[string, User]` | `string` | `entity.User` | L64, L250 | `entity.NewUserRepo(redisStorage UserStorage, ...)` |
| `storageNewRedisStorageStorageCacheEntryEntityOrderInt` | `RedisStorage[int, CacheEntry[Order]]` | `int` | `storage.CacheEntry[entity.Order]` | L65, L256 | `storage.New[int, entity.Order](cache)` |
| `storageNewRedisStorageStorageCacheEntryEntityUserString` | `RedisStorage[string, CacheEntry[User]]` | `string` | `storage.CacheEntry[entity.User]` | L66, L262 | `storage.New[string, entity.User](cache)` |

File: `internal/tests/storage/redis_storage.go:20`

### 3b. `storage.NewDbStorage[V any, K string \| int](conn *DbConnection) *DbStorage[K, V]`

| di.go field | Instantiation | K | V | Line | Consumer |
|------------|--------------|---|---|------|----------|
| `storageNewDbStorageEntityOrderInt` | `DbStorage[int, Order]` | `int` | `entity.Order` | L62, L238 | `entity.NewOrderRepo(dbStorage OrderStorage, ...)` |

File: `internal/tests/storage/db_storage.go:22`

### 3c. `storage.New[K string | int, V any](storage *RedisStorage[K, CacheEntry[V]]) *Cache[K, V]`

| di.go field | Instantiation | K | V | Line | Consumer |
|------------|--------------|---|---|------|----------|
| `storageNewIntEntityOrder` | `Cache[int, Order]` | `int` | `entity.Order` | L67, L268 | `entity.NewOrderRepo(..., cache OrderCache)` |
| `storageNewStringEntityUser` | `Cache[string, User]` | `string` | `entity.User` | L68, L274 | `entity.NewUserRepo(..., cache UserCache)` |

File: `internal/tests/storage/cache.go:13`

## 4. Parameter Name Resolution: userRepo / orderRepo

`entity.UserRepo` and `entity.OrderRepo` — two different providers resolved by different constructor parameter names in consumers.

| Handler package | File | Param | Provider | di.go line |
|----------------|------|-------|----------|------------|
| `handler/user/get` | `user/get/handler.go:21` | `userRepo` | `entity.NewUserRepo` | L196 |
| `handler/user/create` | `user/create/handler.go:22` | `userRepo` | `entity.NewUserRepo` | L184 |
| `handler/user/update` | `user/update/handler.go:20` | `userRepo` | `entity.NewUserRepo` | L202 |
| `handler/user/delete` | `user/delete/handler.go:21` | `userRepo` | `entity.NewUserRepo` | L190 |
| `handler/order/get` | `order/get/handler.go:20` | `orderRepo` | `entity.NewOrderRepo` | L172 |
| `handler/order/create` | `order/create/handler.go:20` | `orderRepo` | `entity.NewOrderRepo` | L160 |
| `handler/order/update` | `order/update/handler.go:29` | `orderRepo` | `entity.NewOrderRepo` | L178 |
| `handler/order/delete` | `order/delete/handler.go:19` | `orderRepo` | `entity.NewOrderRepo` | L166 |

Files: `internal/tests/entity/user.go:32` (`NewUserRepo(redisStorage UserStorage, cache UserCache)`), `internal/tests/entity/order.go:31` (`NewOrderRepo(dbStorage OrderStorage, cache OrderCache)`)

## 5. Storage Resolution by Name: dbStorage / redisStorage

| Consumer | File | Param | Provider | di.go line |
|----------|------|-------|----------|------------|
| `entity.NewOrderRepo(dbStorage OrderStorage, ...)` | `entity/order.go:31` | `dbStorage` | `storage.NewDbStorage` → `storageNewDbStorageEntityOrderInt` | L238 |
| `entity.NewUserRepo(redisStorage UserStorage, ...)` | `entity/user.go:32` | `redisStorage` | `storage.NewRedisStorage` → `storageNewRedisStorageEntityUserString` | L250 |

`UserStorage` and `OrderStorage` — different interfaces in `entity/user.go:20` and `entity/order.go:13`, both with `Get`, `Set`, `Delete`. Resolved by name + type combination.

## 6. Interface Embedding and Narrow Interfaces

### 6a. Struct Embedding: Handler embeds handler.Base

All handler structs embed `handler.Base` to get `Path()` method, implementing `handler.Handler` interface:

```go
type Handler struct {
    handler.Base    // embed Path() from handler.go
    repo Repo       // narrow interface depends on package
}
```

| Handler | File | di.go field |
|---------|------|-------------|
| `handler.dispatcher.Handler` | `handler/dispatcher/handler.go:11` | `handlerDispatcherNew` L47 |
| `handler.admin.stats.Handler` | `handler/admin/stats/handler.go:15` | `handlerAdminStatsNew` L46 |
| `handler.metrics.Handler` | `handler/metrics/handler.go:16` | `handlerMetricsNew` L48 |
| `handler.order.*.Handler` (4) | `handler/order/*/handler.go:15` | L49-L52 |
| `handler.user.*.Handler` (4) | `handler/user/*/handler.go:16` | L53-L56 |

Base file: `internal/tests/handler/handler.go:4`

### 6b. Interface Composition via Embedding

`handler/order/update/handler.go:19-22`:

```go
type OrderReader interface { Get(id int) (entity.Order, bool) }
type OrderWriter interface { Set(int, entity.Order) }
type Repo interface { OrderReader; OrderWriter }  // embedding composition
```

### 6c. Import Alias Interface Embedding

`handler/metrics/handler.go:12-14`:

```go
import testLogger "github.com/diplexhq/diplex/internal/tests/logger"
type Logger interface { testLogger.Logger }
```

Resolved to `loggerNewLogger` (di.go:57, L208).

### 6d. Narrow interface at point of use with composite types: dispatcher.ComplexEventWithComposite

`dispatcher` defines a narrow interface entirely on-site — no embedding, with `any` and `rune` composite types in method signatures:

```go
type ComplexEventWithComposite interface {
    StreamOutputs(context.Context) chan<- event.Event
    StreamInputs(context.Context) <-chan event.ComplexResult
    Process(context.Context, []any, map[string][]rune) (map[rune][]any, []event.PayloadEntry)
}
```

`*event.ComplexEvent` implements all 3 methods of the narrow interface. Resolver matches provider → consumer via interface subset matching.

| Method | Interface signature (dispatcher) | Impl signature (event) |
|--------|---------------------------------|----------------------|
| `Process` | `Process(ctx, []any, map[string][]rune) (map[rune][]any, []PayloadEntry)` | `Process(ctx, []interface{}, map[string][]byte) (map[uint8][]interface{}, []PayloadEntry)` |

Parsing and matching of composite types with built-in aliases:
- `any` ↔ `interface{}` — identical type, `any` is alias in `go/types`
- `rune` ↔ `int32` — identical type, `rune` is alias in `go/types`
- `byte` ↔ `uint8` — identical type, `byte` is alias in `go/types`
- `map[string][]rune` ↔ `map[string][]byte` — different underlying types but same composite structure (map → slice), matched by signature arity and type structure
- `map[rune][]any` ↔ `map[uint8][]interface{}` — matched via `rune`/`uint8` + `any`/`interface{}` alias resolution

di.go L143-148: `handlerDispatcher.New(fn, ch, complex)` where `complex` resolved to `eventNewComplexEvent`.

Files: `internal/tests/handler/dispatcher/handler.go:13-17` (interface), `internal/tests/event/event.go:36-45` (impl)

## 7. Type Aliases and Narrow Interface Aliases

| Alias | Declaration File | Line | Used in |
|-------|-----------------|------|---------|
| `UserAlias = User` | `entity/user.go:6` | L6 | `UserStorage.Get` return type |
| `OrderAlias = Order` | `entity/order.go:11` | L11 | `OrderStorage.Get` return type |
| `User = entity.User` | `handler/user/get/handler.go:10` | L10 | narrow interface `Repo { Get(id string) (User, bool) }` |
| `User = entity.User` | `handler/user/create/handler.go:11` | L11 | narrow interface `Repo { Set(string, User) }` |
| `RedisDsn = Dsn` | `config/config.go:7` | L7 | alias on `config.Dsn`, resolved as same type |

## 8. Slice Aggregation: []handler.Handler

`handler.NewHTTPServer(handlers []Handler)` collects all 11 handlers implementing `handler.Handler` interface (`Path()` + `ServeHTTP()`):

di.go L116-131:

```
[]handler.Handler{
    admin/stats.New(),       // []Repo slice aggregation
    dispatcher.New(),        // func type + chan + interface deps
    metrics.New(),           // named primitives + error provider + import alias
    order/create.New(),      // narrow interface + orderRepo name match
    order/delete.New(),      // narrow interface + orderRepo name match
    order/get.New(),         // narrow interface + orderRepo name match
    order/update.New(),      // interface composition via embedding + orderRepo
    user/create.New(),       // narrow interface + userRepo name match
    user/delete.New(),       // narrow interface + userRepo name match
    user/get.New(),          // narrow interface + userRepo name match
    user/update.New(),       // narrow interface + userRepo name match
}
```

File: `internal/tests/handler/server.go:15`

### 8a. Combinatorial Generic Constraint Widening: []Repo in admin/stats

`handler/admin/stats/handler.go:20`: `func New(repos []Repo)` — collects `UserRepo` and `OrderRepo` via narrow `Repo { Stats(); Slug() }` interface.

Both types come from generic provider chains: `NewUserRepo(redisStorage UserStorage, cache UserCache)` and `NewOrderRepo(dbStorage OrderStorage, cache OrderCache)` where `UserStorage`/`OrderStorage` are satisfied by `RedisStorage` and `DbStorage` through generic constraint widening (`K ∈ {string, int}`, `V ∈ {User, Order}`). The resolver enumerates all generic instantiations, finds two concrete types that both implement the narrow `Repo` interface, and aggregates them into `[]Repo`.

di.go L134-140:
```go
di.handlerAdminStatsNew = sync.OnceValue(func() *handlerAdminStats.Handler {
    return handlerAdminStats.New(
        []handlerAdminStats.Repo{
            di.entityNewOrderRepo(),
            di.entityNewUserRepo(),
        },
    )
})
```

This is a combinatorial scenario: generic providers (`NewOrderRepo`, `NewUserRepo`) → constraint widening (`K ∈ {string, int}`) → two concrete types → both satisfy the narrow `Repo { Stats(); Slug() }` interface → aggregated into `[]Repo`. Unusual but supported pattern.

## 10. Channel Types in event

`event.NotifyChan chan Event` — channel type as a root provider with no dependencies.

di.go L44, L112-114:
```go
eventNewNotifyChan func() event.NotifyChan  // chan Event
```

Used in `dispatcher.New(fn, ch event.NotifyChan, complex)` → `internal/tests/handler/dispatcher/handler.go:26`

## 11. Provider with Error

`metrics.NewHealthChecker(dbDsn, redisDsn config.Dsn, port Port) (*HealthChecker, error)`:

- di.go L58, L212-222
- Error wrap: `if err != nil { panic(err.Error()) }`
- Two `config.Dsn` params resolved by name: `dbDsn` → `configNewDbDsn` (named return), `redisDsn` → `configNewRedisDsn` (alias `RedisDsn = Dsn`)
- File: `internal/tests/metrics/metrics.go:27`

## 12. Composite Types in Narrow Interface Process Method

`ComplexEventWithComposite.Process` в `dispatcher/handler.go:16` — один метод интерфейса принимает и возвращает composite типы (`[]any`, `map[string][]rune`, `map[rune][]any`, `[]PayloadEntry`). Реализация в `event/event.go:44` использует эквивалентные алиасированные типы (`[]interface{}`, `map[string][]byte`, `map[uint8][]interface{}`).

| Param | Interface type | Impl type | Alias |
|-------|---------------|-----------|-------|
| arg2 | `[]any` | `[]interface{}` | `any` = `interface{}` |
| arg3 | `map[string][]rune` | `map[string][]byte` | `rune` = `int32`, `byte` = `uint8` |
| ret1 | `map[rune][]any` | `map[uint8][]interface{}` | `rune`→`uint8` (both int32/uint8), `any` = `interface{}` |
| ret2 | `[]PayloadEntry` | `[]PayloadEntry` | identical |
