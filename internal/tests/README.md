# Test DI Matrix

Full dependency matrix covering all supported DI scenarios.

## Coverage

| # | Scenario | File |
|---|----------|------|
| 1 | Root provider | `config/config.go:NewConfig` |
| 2 | Single-level dep | `config/config.go:NewDBConfig` |
| 3 | Struct methods (RealType) | `config/config.go:Config.DB()` |
| 4 | Generic provider | `entity/repo.go:NewRepo[T]` |
| 5 | Constraint → method match | `entity/repo.go` + `config/config.go` |
| 6 | Narrow interface per consumer | `handler/*/handler.go:Repo` |
| 7 | Slice collection | `handler/server.go:[]Handler` |
| 8 | Struct embedding | `handler/user/get/Handler` embeds `handler.Base` |
| 9 | Interface subset matching | `Repo.Get` ⊂ `Repo[T].Get` |
| 10 | Generic type substitution | `T=Order` in method signatures |
| 11 | Import alias uniqueness | `handlerOrderCreate`, `handlerUserGet`, etc. |
| 12 | Cross-package wiring | 27 providers, 15 handler deps |
| 13 | Primitive type exclusion | `NewBase(path string)` not resolved by DI |
| 14 | DI facade interface | `di/di.go:DI.HttpServer()` |
| 15 | Aliased narrow interface | `handler/user/lookup:User=entity.User` |
| 16 | Nested alias chain | `handler/order/detail:Order=entity.OrderAlias` |
| 17 | Mixed aliased generics | `handler/admin/stats:UserStatsRepo+OrderStatsRepo` |
| 18 | Slice with alias handlers | 15 handlers in `[]handler.Handler` |
| 19 | Multiple concrete impls | First alphabetically |
| 20 | Value type provider | `config/NewDBConfig` returns `DBConfig` by value |
| 21 | 2+ dependency constructor | `dispatcher.New(fn, ch, complex)` |
| 22 | Function type dependency | `callback.HandlerFunc` + `dispatcher.New` |
| 23 | Named struct (non-pointer) | `config/DBConfig` returned by value |
| 24 | Channel type dependency | `event.NotifyChan` + `dispatcher.New` |
| 25 | Generic 2 type params | `cache.NewCache[V, K]` (reversed order) |
| 26 | Map dependency | interface method-based registration |
| 27 | Provider with error return | `metrics/HealthChecker` → panic on build |
| 28 | Named primitive type | `metrics/Port=int`, `Timeout=Duration` |
| 29 | Factory pattern | `metrics/FactoryProvider` returns `HandlerFunc` |