# DI Сценарии

Полный каталог всех поддерживаемых DI паттернов.

## Паттерны провайдеров

| # | Паттерн | Пример |
|---|---------|--------|
| 1 | Корневой провайдер (без deps) | `func NewConfig() *Config` |
| 2 | Одноуровневая зависимость | `func NewDBConfig(cfg *Config) DBConfig` |
| 3 | Метод структуры как провайдер | `func (c *Config) DB() DBConfig` |
| 4 | Generic провайдер | `func NewRepo[T any]() *Repo[T]` |
| 5 | Matching generic constraints | `func NewRepo[T User | Order]() *Repo[T]` |
| 6 | Generic 2 type params | `func NewCache[V, K]() *Cache[V, K]` |
| 7 | Provider с error | `func NewHealthChecker(...) (*HealthChecker, error)` |
| 8 | Named primitive types | `type Port int; func NewPort() Port` |
| 9 | Value type provider | `func NewDBConfig(...) DBConfig` |
| 10 | Factory pattern | `type FactoryProvider func(...) Handler` |
| 11 | Multiple concrete impls | First alphabetically by ID |

## Паттерны интерфейсов и типов

| # | Паттерн | Пример |
|---|---------|--------|
| 12 | Узкий интерфейс на потребителя | `type Repo interface { Get(...) }` |
| 13 | Slice collection | `[]handler.Handler` агрегируются |
| 14 | Struct embedding | `type Handler struct { Base }` |
| 15 | Interface composition via embedding | `type OrderService interface { OrderReader; OrderWriter }` |
| 16 | Import alias (package) | `import testLogger "github.com/..."` |
| 17 | Aliased narrow interface | `type User = entity.User` |
| 18 | Nested alias chain | `type Order = entity.OrderAlias` |
| 19 | Mixed aliased generics | `type UserStatsRepo = Repo[User]` |
| 20 | Import alias interface embedding | `import api "..."` + `type Manager interface { api.Store }` |
| 21 | Pointer embed в структуре | `type Tracker struct { *domain.Timestamps }` |
| 22 | Interface subset matching | `Repo.Get` ⊂ `Repo[T].Get` |

## Продвинутые паттерны типов

| # | Паттерн | Пример |
|---|---------|--------|
| 23 | Chan direction types | `chan<- Item`, `<-chan []Item` |
| 24 | Map types в методах | `ListItems(filter map[string]string)` |
| 25 | Variadic методы | `Sum(items ...*Item) error` |
| 26 | Named return values | `Report() (count int, total float64, err error)` |
| 27 | Cross-service dependency chain | `Service` → `Repo` + `Gateway` + `Validator` |
| 28 | Constructor с 4+ deps | `NewService(a, b, c, d)` |
| 29 | Function type dependency | `callback.HandlerFunc` |
| 30 | Channel type dependency | `event.NotifyChan` |
| 31 | Any/interface{} normalization | `type FilterArgs any` |
| 32 | Cross-module dependency chain | `OrderService` → `PaymentService` → `StripeGateway` |
| 33 | Nested generic instantiation | `Repository[T]` → `InMemoryStore[Order]` → `Service` |

## Resolution и Filtering

| # | Паттерн | Описание |
|---|---------|----------|
| 34 | Private type filtering | Private structs (`lowercaseName`) исключены |
| 35 | DI facade interface | `di.DI` интерфейс с методами |
| 36 | Handler slice aggregation | Все handlers собраны в `[]handler.Handler` |

## Ограничения

### Нерешённые проблемы

- **Пересечение узких интерфейсов**: DI-ресолвер не умеет различать узкие интерфейсы разных потребителей с одинаковыми методами. Если два handler-а требуют `Repo.Get(id int) (entity.Order, error)` и `Repo.Get(id int) (entity.User, error)` — конфликт неразрешим. В таких случаях необходимо использовать широкие интерфейсы (типа `entity.Repository`).
- **Сложные инлайн типы**: `map[string]string`, `chan<- Event`, `<-chan ComplexResult` и другие сложные composite типы в сигнатурах методов — не цель авто-резолвера. Рекомендуется определять именованные типы и алиасы (например, `type FilterArgs = map[string]string`) в отдельных пакетах.

### Правила сигнатур провайдеров

| Правило | Поведение |
|---------|----------|
| Максимум 2 return values | Конструкторы с 3+ return values **игнорируются** |
| Optional error | Второй return value должен быть `error` |
| Pointer или value | Первый return должен быть `*T` или `T` |

### Правила generic типов

| Правило | Поведение |
|---------|----------|
| Type param normalization | `T0`, `T1` → выровнены до `T`, `T0`, `T1`... |
| Constraint widening | `[T User \| Order]` → concrete instantiations для каждого constraint |
| Empty constraints | Generic без constraints **не производит** instantiations |

## Интеграционные тесты

Все сценарии покрыты тестами в `internal/tests/`:

```
internal/tests/
├── cache/         — Generic типы, nested instantiations
├── config/        — Root провайдеры, named primitives
├── di/            — DI facade interface
├── entity/        — Generic providers, narrow interfaces
├── handler/       — Slice aggregation, struct embed
├── logger/        — Private type filtering
└── metrics/       — 4+ dependency constructor
```

**Источник истины:** `internal/tests/di_test.go`.