# Архитектура проекта

## Обзор пайплайна

```
go/parser → AST → Scanner → chan SourceFile → Parser → ParsedData → Resolver → ResolvedData → Generator → di.go
```

Пайплайн состоит из четырёх этапов: **scan → parse → resolve → generate**. Каждый этап produces well-defined промежуточное представление, которое потребляется следующим.

## Этап 1: Scanner (`internal/scanner/`)

Один горутин проходит по директориям через `filepath.WalkDir`, применяет фильтры skip pattern и отправляет принятые `.go` пути в буферизированный канал (`chan SourceFile`, capacity 4). Параллелизм отложен до этапа парсера.

## Этап 2: Parser (`internal/parser/`)

4 горутина потребляют канал параллельно, парся Go AST и извлекая провайдеры (`New*` функции), интерфейсы, реализации (methods receiver), type aliases и embeds. Изменяемое состояние защищено `sync.Mutex`. Пост-парсинг: алиасы flatten, embeds BFS-flatten.

## Этап 3: Resolver (`internal/resolver/`)

Строит два индекса (`typeIndex` + `interfaceMethodIndex`) из ParsedData и разрешает зависимости через BFS traversal от facade method result types. Поддерживает generic constraint narrowing и combinatorial resolution.

> **Глубокие детали** — алгоритмы, сужение констрейнов, комбинаторика, характеристики производительности: см. [ARCHITECTURE.md](ARCHITECTURE.md).

## Этап 4: Generator (`internal/generator/`)

Генерирует DI-контейнер `.go` файлы из `ResolvedData` используя четыре embedded шаблона (`head`, `import`, `facade`, `provider`). Provider поля обёрнуты в `sync.OnceValue()` для lazy initialization. Output детерминирован (SHA-256 verified).

## Принципы дизайна

- **Interface-based DI**: Зависимости wire через интерфейсы, определённые в `-di` директориях (facades). Сгенерированный код имплементирует эти facades.
- **Narrow interfaces at point of use**: Каждый consumer определяет minimal interface. Resolver matches implementations через method signature comparison.
- **Name-based disambiguation**: Когда multiple providers satisfy type, parameter names resolve ambiguity через `ResultName` matching.
- **Slice aggregation**: `[]T` parameters автоматически collect all matching providers.
- **Ноль внешних зависимостей**: Только стандартная библиотека (`go/parser`, `go/types`, `text/template`, `sync`).

## Пайплайн разрешения type aliases

```
parseTypeAliases → resolveAliases → resolveProviders → resolveMethods
                    │                │                  │
                    │                │                  └─ replace Params & Results
                    │                └─ replace Arguments & Result
                    └─ BFS-flatten alias chains
```

Built-in aliases (`byte`→`uint8`, `rune`→`int32`, `interface{}`→`any`) resolved в `parseTypeAliases`, затем applied transitively ко всем provider arguments, results и method signatures во время `resolveAliases`.

## Matching method signatures

Resolver matches interface methods через two-tier key lookup в `interfaceMethodIndex`:
1. **Full signature key** — `"Name(args)(results)"` для non-generic методов
2. **Bare name key** — `"Name"` для generic методов

Оба queried, results unioned. Это гарантирует что и concrete implementations и generic providers satisfy interface requirements.
