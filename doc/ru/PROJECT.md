# DIplex — генератор DI-контейнера для Go

## Назначение

Автоматическое сканирование Go-кода и генерация высокопроизводительного DI-контейнера на этапе кодгенерации (code generation).

## Основные принципы

- **Compile-time DI** — без runtime reflection, максимальная производительность
- **Type-safe** — все зависимости проверяются при компиляции
- **Zero overhead** — сгенерированный код равен ручному внедрению зависимостей
- **Convention over configuration** — использует стандартные паттерны Go

## Как это работает

### 1. Сканирование
Рекурсивный обход Go-файлов в заданных директориях. Пропускает `_test.go` и директории `mocks/`. Парсинг AST через стандартный `go/parser`.

### 2. Анализ
Извлекает три типа сущностей:

**Провайдеры:** функции с паттерном `New*`, возвращающие указатель (опционально с `error`). Поддерживают дженерики: `NewRepo[T User | Order]` производит concrete instantiations.

**Интерфейсы:** объявленные интерфейсы с методами и встроенными интерфейсами. Поддерживают дженерики: `Repository[T any]`.

**Реализации:** публичные методы структур, включая generic receiver: `func (r *Repo[Order]) Get(id int) (Order, error)`.

### 3. Генерация
Создает DI-фасады в выходной директории. Каждая `-di` директория производит один `.go` файл со структурой `DI` и конструктором `NewDI()`, использующим `sync.OnceValue` для singleton-семантики.

Имена полей для generic-провайдеров санитизируются в camelCase: `entity.NewRepo[entity.Order]` → `entityNewRepoEntityOrder`.

## Структура проекта

```
├── main.go                    # Точка входа, оркестрация
└── internal/
    ├── generator/             # Генерация кода (шаблоны: facade, provider, head, import)
    ├── resolver/              # Разрешение провайдеров и маппинг зависимостей
    │   ├── build_index.go     # typeIndex + interfaceMethodIndex
    │   ├── find_providers.go  # Поиск провайдеров и matching constraints
    │   └── resolve*.go        # BFS обход зависимостей
    ├── scanner/               # Сканирование Go файлов
    ├── parser/                # Парсинг и анализ AST
    │   ├── ast_stringer/      # AST → string (каноническое форматирование)
    │   └── resolve*.go        # Алиасы и embeds
    ├── domain/                # Доменные типы (ParsedData, Provider, MethodContract)
    ├── di/                    # Интерфейс DI-фасада
    ├── config/                # CLI-флаги и конфигурация
    └── utils/                 # Must, NoErr, SanitizeIdent, Logger
```

## Использование

```bash
# Установить как Go-утилиту
go get github.com/diplexhq/diplex

# Запуск из корня целевого проекта (автоматически читает go.mod)
go tool diplex

# Сканировать конкретные директории
go tool diplex -scan internal,pkg

# Пользовательская директория вывода
go tool diplex -out internal/generated/diplex

# Явный путь модуля (без чтения go.mod)
go tool diplex -module example.com/my/project

# Пользовательский паттерн пропуска (regexp)
go tool diplex -skip "(internal\/generated\/diplex|testdata|mocks?|_test\.go|_mock\.go)$"

# Указать директории с DI-фасадами
go tool diplex -di internal/di

# Подробный / тихий режим
go tool diplex -v
go tool diplex -s
```

**Умолчания:** сканирует `internal/`, DI-фасады в `internal/di/`, вывод в `internal/generated/diplex/`.

## Ограничения

- **Примитивные типы должны быть обёрнуты** — `string`, `int` и др. не могут быть сопоставлены. Каждый примитивный аргумент должен использовать уникальный именованный тип.
- **Узкие интерфейсы** — объявляйте интерфейсы по месту использования, не создавайте широкие глобальные интерфейсы.
- **Внешние интерфейсы** — сканируются только интерфейсы из project code, не из stdlib или внешних пакетов.
- **Provider возвращает 1-2 значения** — поддерживаются только `T`, `T, error`, `*T`, `*T, error`.
- **Все провайдеры — синглтоны** через `sync.OnceValue`. Циклические зависимости вызывают panic при запуске.

## Стиль интерфейсов

```go
// ХОРОШО — узкий интерфейс потребителя
type ScanDirProvider interface {
    ScanDirs() string
    Silent() bool
}

// ПЛОХО — широкий интерфейс
type GlobalConfig interface {
    ScanDirs() string
    OutputDir() string
    IsSilent() bool
}
```

Избегайте префиксов `Get` и `Is` в именах методов. Используйте существительные или прилагательные напрямую.

## Полное покрытие сценариев

Смотрите [SCENARIOS.md](SCENARIOS.md) — полный каталог всех поддерживаемых DI сценариев и интеграционных тестов в `internal/tests/`.