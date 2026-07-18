# DI Сценарии

Полный каталог всех DI-паттернов, покрытых `internal/tests/` и верифицированных сгенерированным контейнером в `internal/tests/generated/diplex/di.go`.

## 1. Корневые провайдеры (без зависимостей)

Функции `New*()` без параметров.

| Источник | di.go поле | Строка в di.go | Тип |
|----------|-----------|----------------|-----|
| `config.NewConfig() *Config` | `configNewConfig` | L74 | `*config.Config` |
| `logger.NewLogger() Logger` | `loggerNewLogger` | L208 | `logger.Logger` |
| `event.NewNotifyChan() NotifyChan` | `eventNewNotifyChan` | L112 | `chan Event` (channel) |
| `event.NewHandlerFunc() HandlerFunc` | `eventNewHandlerFunc` | L108 | `func(path string) string` (function type) |
| `event.NewComplexEvent() *ComplexEvent` | `eventNewComplexEvent` | L104 | `*event.ComplexEvent` |
| `metrics.NewPort() Port` | `metricsNewPort` | L224 | `int` (named primitive) |
| `metrics.NewTimeout() Timeout` | `metricsNewTimeout` | L228 | `time.Duration` (named primitive) |
| `storage.NewDbConnection(Dsn) *DbConnection` | `storageNewDbConnection` | L61, L232 | `*storage.DbConnection` (используется generic-ами ниже) |

Файлы: `internal/tests/config/config.go:17`, `internal/tests/logger/logger.go:16`, `internal/tests/event/event.go:36`, `internal/tests/event/callback.go:6`, `internal/tests/metrics/metrics.go:13`, `internal/tests/metrics/metrics.go:19`, `internal/tests/storage/db_storage.go:11`

## 2. Провайдеры с одноуровневой зависимостью

| Источник | di.go поле | Строка | Зависимость | Тип результата |
|----------|-----------|--------|-------------|---------------|
| `config.NewDbDsn(*Config) (dbDsn Dsn)` | `configNewDbDsn` | L38, L78 | `configNewConfig` | `config.Dsn` (named `string`) |
| `config.NewRedisDsn(*Config) RedisDsn` | `configNewRedisDsn` | L39, L84 | `configNewConfig` | `config.RedisDsn` (alias `= Dsn`) |
| `storage.NewRedisConnection(redisDsn Dsn) *RedisConnection` | `storageNewRedisConnection` | L63, L244 | `configNewRedisDsn` | `*storage.RedisConnection` |
| `storage.NewDbConnection(dbDsn Dsn) *DbConnection` | `storageNewDbConnection` | L61, L232 | `configNewDbDsn` | `*storage.DbConnection` |

Файлы: `internal/tests/config/config.go:25,29`, `internal/tests/storage/redis_storage.go:9`, `internal/tests/storage/db_storage.go:11`

### Разрешение DSN по имени

`metrics.NewHealthChecker(dbDsn, redisDsn config.Dsn, port Port)` — два параметра типа `config.Dsn` разрешаются по имени параметра через именованный результат провайдера:

- `dbDsn` → `config.NewDbDsn(...) (dbDsn Dsn)` — named return value `dbDsn` в `config/config.go:29`
- `redisDsn` → `storage.NewRedisConnection(redisDsn config.Dsn)` — named параметр в `storage/redis_storage.go:9`
- `config.RedisDsn = Dsn` — type alias в `config/config.go:7`, резолвится как тот же тип `config.Dsn`

## 3. Generic провайдеры: цепочки с constraint widening

### 3a. `storage.NewRedisStorage[V any, K string \| int](conn *RedisConnection) *RedisStorage[K, V]`

Signature params: `[V any, K string | int]` — обратный порядок в instantiation: `NewRedisStorage[User, string]` при signature `[V, K]`.

| di.go поле | Instantiation | K | V | Строка | Потребитель |
|-----------|--------------|---|---|--------|------------|
| `storageNewRedisStorageEntityUserString` | `RedisStorage[string, User]` | `string` | `entity.User` | L64, L250 | `entity.NewUserRepo(redisStorage UserStorage, ...)` |
| `storageNewRedisStorageStorageCacheEntryEntityOrderInt` | `RedisStorage[int, CacheEntry[Order]]` | `int` | `storage.CacheEntry[entity.Order]` | L65, L256 | `storage.New[int, entity.Order](cache)` |
| `storageNewRedisStorageStorageCacheEntryEntityUserString` | `RedisStorage[string, CacheEntry[User]]` | `string` | `storage.CacheEntry[entity.User]` | L66, L262 | `storage.New[string, entity.User](cache)` |

Файл: `internal/tests/storage/redis_storage.go:20`

### 3b. `storage.NewDbStorage[V any, K string \| int](conn *DbConnection) *DbStorage[K, V]`

| di.go поле | Instantiation | K | V | Строка | Потребитель |
|-----------|--------------|---|---|--------|------------|
| `storageNewDbStorageEntityOrderInt` | `DbStorage[int, Order]` | `int` | `entity.Order` | L62, L238 | `entity.NewOrderRepo(dbStorage OrderStorage, ...)` |

Файл: `internal/tests/storage/db_storage.go:22`

### 3c. `storage.New[K string | int, V any](storage *RedisStorage[K, CacheEntry[V]]) *Cache[K, V]`

| di.go поле | Instantiation | K | V | Строка | Потребитель |
|-----------|--------------|---|---|--------|------------|
| `storageNewIntEntityOrder` | `Cache[int, Order]` | `int` | `entity.Order` | L67, L268 | `entity.NewOrderRepo(..., cache OrderCache)` |
| `storageNewStringEntityUser` | `Cache[string, User]` | `string` | `entity.User` | L68, L274 | `entity.NewUserRepo(..., cache UserCache)` |

Файл: `internal/tests/storage/cache.go:13`

## 4. Разрешение по имени параметра: userRepo / orderRepo

`entity.UserRepo` и `entity.OrderRepo` — два разных провайдера, разные интерфейсы, разные имена параметров потребителей.

| Handler пакет | Файл | Параметр | Провайдер | Строка в di.go |
|--------------|------|----------|-----------|----------------|
| `handler/user/get` | `user/get/handler.go:21` | `userRepo` | `entity.NewUserRepo` | L196 |
| `handler/user/create` | `user/create/handler.go:22` | `userRepo` | `entity.NewUserRepo` | L184 |
| `handler/user/update` | `user/update/handler.go:20` | `userRepo` | `entity.NewUserRepo` | L202 |
| `handler/user/delete` | `user/delete/handler.go:21` | `userRepo` | `entity.NewUserRepo` | L190 |
| `handler/order/get` | `order/get/handler.go:20` | `orderRepo` | `entity.NewOrderRepo` | L172 |
| `handler/order/create` | `order/create/handler.go:20` | `orderRepo` | `entity.NewOrderRepo` | L160 |
| `handler/order/update` | `order/update/handler.go:29` | `orderRepo` | `entity.NewOrderRepo` | L178 |
| `handler/order/delete` | `order/delete/handler.go:19` | `orderRepo` | `entity.NewOrderRepo` | L166 |

Файлы: `internal/tests/entity/user.go:32` (`NewUserRepo(redisStorage UserStorage, cache UserCache)`), `internal/tests/entity/order.go:31` (`NewOrderRepo(dbStorage OrderStorage, cache OrderCache)`)

## 5. Разрешение storage по имени: dbStorage / redisStorage

| Потребитель | Файл | Параметр | Провайдер | Строка в di.go |
|------------|------|----------|-----------|----------------|
| `entity.NewOrderRepo(dbStorage OrderStorage, ...)` | `entity/order.go:31` | `dbStorage` | `storage.NewDbStorage` → `storageNewDbStorageEntityOrderInt` | L238 |
| `entity.NewUserRepo(redisStorage UserStorage, ...)` | `entity/user.go:32` | `redisStorage` | `storage.NewRedisStorage` → `storageNewRedisStorageEntityUserString` | L250 |

`UserStorage` и `OrderStorage` — разные интерфейсы в `entity/user.go:20` и `entity/order.go:13`, но оба имеют методы `Get`, `Set`, `Delete`. Ресолвер различает их по имени параметра + типу.

## 6. Interface embedding и narrow interfaces

### 6a. Struct embedding: Handler embeds handler.Base

Все handler-структуры embed `handler.Base` для получения `Path()` метода, что реализует `handler.Handler` интерфейс:

```go
type Handler struct {
    handler.Base    // embed Path() из handler.go
    repo Repo       // narrow interface зависит от пакета
}
```

| Handler | Файл | di.go поле |
|---------|------|-----------|
| `handler.dispatcher.Handler` | `handler/dispatcher/handler.go:19` | `handlerDispatcherNew` L47 |
| `handler.admin.stats.Handler` | `handler/admin/stats/handler.go:15` | `handlerAdminStatsNew` L46 |
| `handler.metrics.Handler` | `handler/metrics/handler.go:16` | `handlerMetricsNew` L48 |
| `handler.order.*.Handler` (4 шт) | `handler/order/*/handler.go:15` | L49-L52 |
| `handler.user.*.Handler` (4 шт) | `handler/user/*/handler.go:16` | L53-L56 |

Файл base: `internal/tests/handler/handler.go:4`

### 6b. Interface composition via embedding

`handler/order/update/handler.go:19-22`:

```go
type OrderReader interface { Get(id int) (entity.Order, bool) }
type OrderWriter interface { Set(int, entity.Order) }
type Repo interface { OrderReader; OrderWriter }  // embedding композиция
```

### 6c. Import alias interface embedding

`handler/metrics/handler.go:12-14`:

```go
import testLogger "github.com/diplexhq/diplex/internal/tests/logger"
type Logger interface { testLogger.Logger }
```

Резолвится в `loggerNewLogger` (di.go:57, L208).

### 6d. Narrow interface по месту использования с composite типами: dispatcher.ComplexEventWithComposite

`dispatcher` определяет узкий интерфейс полностью на месте — без embedding, с `any` и `rune` composite типами в сигнатурах:

```go
type ComplexEventWithComposite interface {
    StreamOutputs(context.Context) chan<- event.Event
    StreamInputs(context.Context) <-chan event.ComplexResult
    Process(context.Context, []any, map[string][]rune) (map[rune][]any, []event.PayloadEntry)
}
```

`*event.ComplexEvent` имплементирует все 3 метода narrow interface. Ресолвер сопоставляет provider → consumer через interface subset matching.

| Метод | Интерфейс (dispatcher) | Имплеметация (event) |
|-------|----------------------|---------------------|
| `Process` | `Process(ctx, []any, map[string][]rune) (map[rune][]any, []PayloadEntry)` | `Process(ctx, []interface{}, map[string][]byte) (map[uint8][]interface{}, []PayloadEntry)` |

Парсинг и сопоставление composite типов с учётом встроенных алиасов:
- `any` ↔ `interface{}` — один тип, алиас из `go/types`
- `rune` ↔ `int32` — один тип, алиас из `go/types`
- `byte` ↔ `uint8` — один тип, алиас из `go/types`
- `map[string][]rune` ↔ `map[string][]byte` — разные типы, но ресолвер сопоставляет по сигнатуре метода (совпадают по arity и базовой структуре composite)
- `map[rune][]any` ↔ `map[uint8][]interface{}` — аналогично через алиасы

di.go L143-148: `handlerDispatcher.New(fn, ch, complex)` где `complex` резолвится в `eventNewComplexEvent`.

Файлы: `internal/tests/handler/dispatcher/handler.go:13-17` (interface), `internal/tests/event/event.go:36-45` (impl)

## 7. Type aliases и алиасы в узких интерфейсах

| Alias | Файл объявления | Строка | Используется в |
|-------|----------------|--------|---------------|
| `UserAlias = User` | `entity/user.go:6` | L6 | `UserStorage.Get` return type |
| `OrderAlias = Order` | `entity/order.go:11` | L11 | `OrderStorage.Get` return type |
| `User = entity.User` | `handler/user/get/handler.go:10` | L10 | narrow interface `Repo { Get(id string) (User, bool) }` |
| `User = entity.User` | `handler/user/create/handler.go:11` | L11 | narrow interface `Repo { Set(string, User) }` |
| `RedisDsn = Dsn` | `config/config.go:7` | L7 | alias на `config.Dsn`, резолвится как один тип |

## 8. Slice агрегация: []handler.Handler

`handler.NewHTTPServer(handlers []Handler)` собирает все 11 handlers, имплементирующих `handler.Handler` интерфейс (`Path()` + `ServeHTTP()`):

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

Файл: `internal/tests/handler/server.go:15`

### 8a. Комбинаторный generic constraint widening: []Repo в admin/stats

`handler/admin/stats/handler.go:20`: `func New(repos []Repo)` — собирает `UserRepo` и `OrderRepo` через narrow интерфейс `Repo { Stats(); Slug() }`.

Оба типа получаются из generic-провайдеров `NewUserRepo` и `NewOrderRepo` с constraint widening через `storage.NewRedisStorage` и `storage.NewDbStorage` — `K ∈ {string, int}` и `V ∈ {User, Order}` генерируют несколько instantiations, оба concrete type удовлетворяют узкому интерфейсу `Repo { Stats(); Slug() }` и агрегируются в `[]Repo`.

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

Это комбинаторный сценарий: generic провайдеры (`NewOrderRepo`, `NewUserRepo`) → constraint widening (`K ∈ {string, int}`) → два concrete type → оба удовлетворяют narrow интерфейсу `Repo { Stats(); Slug() }` → агрегируются в `[]Repo`.

## 9. Provider с error

`metrics.NewHealthChecker(dbDsn, redisDsn config.Dsn, port Port) (*HealthChecker, error)`:

- di.go L58, L212-222
- Error wrap: `if err != nil { panic(err.Error()) }`
- Два `config.Dsn` параметра разрешаются по имени: `dbDsn` → `configNewDbDsn` (named return), `redisDsn` → `configNewRedisDsn` (alias `RedisDsn = Dsn`)
- Файл: `internal/tests/metrics/metrics.go:27`
