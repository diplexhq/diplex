# DIplex — Go DI Generator

Utility for scanning Go code and generating a high-performance DI container at compile time. No runtime reflection — generated code equals manual dependency injection.

# DIplex — генератор DI для Go

Утилита для сканирования Go-кода и генерации высокоэффективного DI-контейнера на этапе компиляции. Без runtime reflection — сгенерированный код не уступает ручному внедрению зависимостей.

- **No deps** / **Нет зависимостей**
- **No config** / **Нет конфигурации**
- **No overhead** / **Нет накладных расходов**

---

## Quick Start / Быстрый старт

### Installation / Установка

```bash
go get -tool github.com/diplexhq/diplex@latest
go mod vendor
```

### 1. Write code as usual / Пишите код как обычно

Define providers with `New*` functions, define a `DI` interface facade.

```go
// Provider — функция New*, возвращает *Type или (Type, error)
func NewConfig() *Config
func NewDB(cfg *Config) (*DB, error)

// DI facade
type DI interface {
    Config() *config.Config
    DB() *db.DB
}
```

### 2. Run the generator / Запустите генератор

```bash
go tool diplex
```

Generated `di.go` appears in the output directory with wired dependencies.

### 3. Use it / Используйте

```go
deps := diplex.NewDI()
cfg := deps.Config()
db := deps.DB()
```

---

## Documentation / Документация

### User-facing / Для пользователей
| | English | Русский |
|---|---------|---------|
| User Guide | [doc/en/USER_GUIDE.md](doc/en/USER_GUIDE.md) | [doc/ru/USER_GUIDE.md](doc/ru/USER_GUIDE.md) |
| Scenarios (feature catalog) | [doc/en/SCENARIOS.md](doc/en/SCENARIOS.md) | [doc/ru/SCENARIOS.md](doc/ru/SCENARIOS.md) |

### Developer-facing / Для разработчиков
| | English | Русский |
|---|---------|---------|
| Developer Guide | [doc/en/DEVELOPER_GUIDE.md](doc/en/DEVELOPER_GUIDE.md) | [doc/ru/DEVELOPER_GUIDE.md](doc/ru/DEVELOPER_GUIDE.md) |
| Architecture | [doc/en/ARCHITECTURE.md](doc/en/ARCHITECTURE.md) | [doc/ru/ARCHITECTURE.md](doc/ru/ARCHITECTURE.md) |
