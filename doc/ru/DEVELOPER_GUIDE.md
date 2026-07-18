# Руководство разработчика

## Принципы

- **Ноль зависимостей** — только стандартная библиотека Go. Никаких внешних пакетов.
- **Разработка через тестирование** — каждая фича должна быть реализована и покрыта в `internal/tests/`. Каждый сценарий — это source-файл с провайдером, узким интерфейсом по месту использования и соответствующим DI-подключением в сгенерированном `di.go`.
- **Самоприменимость** — DIplex генерирует DI-контейнер для самого себя. И `main.go`, и `internal/tests/` используют один и тот же генератор.
- **Минимализм** — без переусложнения. Один concern = один файл.
- **Типобезопасность** — всё, что можно проверить на этапе компиляции, проверяется.
- **Читаемость** — сгенерированный код должен быть самодокументируемым.

## CI / Git-хуки

Git pre-commit hook (`.githooks/pre-commit`) запускается на проиндексированных `.go` файлах:
1. `golangci-lint` — линтинг (автоформатирование через `gofumpt`)
2. `go vet` — статический анализ
3. `go build ./...` — проверка компиляции

Установка: `cp .githooks/pre-commit .git/hooks/pre-commit && chmod +x`

## Технологический стек

- Go 1.26+
- Только стандартная библиотека: `go/parser`, `go/types`, `text/template`, `sync`, `context`, `os`, `io`
- Без reflection, без кодгенерации кроме самого diplex

## Стиль кода

**Импорт алиасы:** имена пакетов в snake_case должны использовать camelCase-алиасы.

```go
// ПРАВИЛЬНО
import astStringer "github.com/diplexhq/diplex/internal/parser/ast_stringer"
```

**Обработка ошибок:** паники — основной механизм. Используйте `panic("message")` для сбоев генерации. Используйте `utils.Must()` и `utils.NoErr()` щедро. `os.Exit()` — только в `main.go` defer recovery.

```go
panic("corrupted Go source: empty identifier name — fix your source code")
```

## Обзор архитектуры

```
main.go
  └─ diplex.NewDI() → DI facade (config, logger, scanner, parser, resolver, generator) — сгенерировано в `internal/generated/diplex/di.go`
        │
        ├─ Scanner.Scan() → domain.SourceFiles (chan SourceFile)
        ├─ Parser.Parse(SourceFiles) → domain.ParsedData (providers + interfaces)
        ├─ Resolver.Resolve(ParsedData) → domain.ResolvedData
        └─ Generator.Generate(ResolvedData) → .go files
```

Каждый этап — отдельный пакет (`internal/scanner/`, `internal/parser/`, `internal/resolver/`, `internal/generator/`). Подробные внутренние механизмы каждого этапа — алгоритмы, структуры данных, сложность — описаны в [ARCHITECTURE.md](ARCHITECTURE.md).

### Ключевые архитектурные решения

- **Scanner** produces buffered channel (`chan SourceFile`, capacity 4). Files walked by single goroutine; parallelism deferred to parsing.
- **Parser** uses exactly 4 worker goroutines sharing a mutex-protected `parseState`. Post-parse, aliases and embeds resolved in sequence.
- **Resolver** builds two-index system (`typeIndex` + `interfaceMethodIndex`), resolves dependencies via BFS from facade methods. Constraint narrowing and combinatorial generic resolution documented in [ARCHITECTURE.md](ARCHITECTURE.md).
- **Generator** renders four templates (`head`, `import`, `facade`, `provider`) into deterministic output verified by SHA-256 hash.

## Тестирование

Все DI-сценарии должны быть реализованы и покрыты в `internal/tests/`. Интеграционный тест `internal/tests/di_test.go` проверяет детерминированность генерации через SHA-256 хеш `di.go`.

Для обновления хеша после изменений генератора:
```bash
go run . -scan internal/tests -di internal/tests/di -out internal/tests/generated/diplex
sha256sum internal/tests/generated/diplex/di.go
# обновить expectedHash в internal/tests/di_test.go
```

## Самопроверка

```bash
# Сборка DI для самого проекта
go run .

# Сборка DI для тестовой матрицы
go run . -scan internal/tests -di internal/tests/di -out internal/tests/generated/diplex
```

Оба должны компилироваться: `go build ./...` должен проходить.
