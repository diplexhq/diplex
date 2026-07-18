# Глубокое погружение в архитектуру

Этот документ описывает внутреннюю архитектуру DIplex подробно. Он охватывает сканирование, параллельность, парсинг, двухиндексную архитектуру ресолвера, сужение констрейнов и комбинаторику дженериков.

## Этапы пайплайна

```
go/parser → AST → Scanner → chan SourceFile → Parser → ParsedData → Resolver → ResolvedData → Generator → di.go
```

Пайплайн состоит из четырёх этапов: **scan → parse → resolve → generate**. Каждый этап produces well-defined промежуточное представление, которое потребляется следующим.

## Этап 1: Сканирование файлов

### Точка входа

`Scanner.Scan()` (`internal/scanner/scan.go`) возвращает `chan SourceFile` (типаlias для `string`, представляющий абсолютный путь к файлу). Буфер канала — 4.

### Логика обхода

Один горутин проходит по всем `scanDirs`. Для каждой директории `filepath.WalkDir()` рекурсивно спускается по дереву. На каждом entry применяются три фильтра:

1. **Пропуск директории** — если `skipPattern.MatchString(path)` для директории, возвращается `filepath.SkipDir`, предотвращая рекурсию.
2. **Фильтр по суффиксу** — не-`.go` файлы пропускаются полностью.
3. **Фильтр по пути** — если `skipPattern.MatchString(path)` для файла, он пропускается.

```
filepath.WalkDir(dir, func(path, DirEntry, error) {
    if d.IsDir() && skipPattern.Match(path) → filepath.SkipDir
    if !strings.HasSuffix(path, ".go") → continue
    if skipPattern.Match(path) → continue
    ch <- SourceFile(path)
})
```

Дефолтный skip pattern: `` `(internal\/generated\/diplex|tests|mocks?|_test\.go|_mock\.go)$` ``. Можно переопределить через флаг `-skip`.

### Модель параллельности

Сканер **однопоточный** — один горутин ходит по всем директориям и шлёт файлы в канал. Параллельность отложена до этапа парсера — 4 потребительских горутин. Это избегает overhead на горутину за файл при минимальном буфере канала (capacity 4).

### Краевые случаи

- `internal/di` не входит в дефолтный scan directory и должен быть явно добавлен через `-scan`.
- Сгенерированные файлы всегда исключены дефолтным skip pattern.
- Канал закрывается после полного обхода всех директорий, сигнализируя EOF потребителям.

## Этап 2: AST Парсинг

### Параллельный парсинг

Парсер создаёт ровно **4 воркер-горутины** (`parser.go:65`). Каждая воркер ranges over канал `SourceFiles` до закрытия, вызывая `parseFile()` для каждого файла.

### Изменяемое состояние

`parseState` (`parser.go:32-40`) хранит все промежуточные данные:

```go
type parseState struct {
    providers     domain.Providers          // ProviderID → *Provider
    interfaces    domain.Interfaces         // InterfaceID → InterfaceInfo
    embeds        map[InterfaceID][]InterfaceID  // interface → embedded interfaces
    typeAliases   map[string]string         // pkg.Alias → pkg.BaseType
    resolvedTypes map[string]string         // cache для alias resolution
    mu            sync.Mutex                // защищает параллельный доступ
    fSet          *token.FileSet            // shared AST file set
}
```

Все мутации `providers`, `interfaces`, `embeds`, `typeAliases`, `resolvedTypes` сериализуются через `state.mu.Lock()`.

### Извлечение сущностей

#### Провайдеры (`parse_provider.go`)

Функция qualifies как DI-провайдер когда ВСЕ условия выполнены:

| Условие | Проверка |
|---------|----------|
| Нет receiver | `funcDecl.Recv == nil` |
| Префикс имени | `strings.HasPrefix(name, "New")` |
| Кол-во результатов | `len(results) == 1 or 2` |
| Первый результат exported | `utils.IsExported(baseIdent)` |
| Optional error | Если 2 результата, второй должен быть `error` |

**Извлечение generic параметров** (`extractGenericParams`):
- Один тип-параметр → alias `T`, suffix `[T]`
- Несколько тип-параметров → aliases `T`, `T1`, `T2`, ... suffix `[T, T1, ...]`
- Union constraints (`User | Order`) → flatten через `walkConstraint()` в `[]string{"User", "Order"}`
- Interface constraints (e.g., `fmt.Stringer`) → single-element list

#### Интерфейсы (`parse_interface.go`)

Для каждого `*ast.InterfaceType`:
- Методы сериализуются через `astStringer.FieldsToStrings()` — **имена параметров discard**, важны только типы
- Embedded interfaces записываются в `state.embeds`
- `RealType: false` — объявленные интерфейсы vs. реализации

#### Методы receiver (`parse_receiver.go`)

Методы на concrete типах становятся имплементациями интерфейсов:
- Exported methods на exported structs
- Receiver может быть pointer или value
- Generic receivers производят `[T]` suffixes и type aliases
- `RealType: true` — concrete типы vs. interface declarations

#### Type Aliases (`parse_type_aliases.go` + `resolve_aliases.go`)

- `type X = Y` declarations populate `state.typeAliases`
- Built-in aliases pre-loaded: `byte→uint8`, `rune→int32`, `interface{}→any`
- `resolveAliases()` flattens alias chains transitively (A→B→C becomes A→C), detects cycles

### Пост-парсинг разрешение

После парсинга всех файлов (все 4 воркера finished):

1. **Alias resolution** (`resolveAliases`) — BFS-flatten alias chains, replace все типы в провайдерах, методах и интерфейсных контрактах
2. **Embed resolution** (`resolveEmbeds`) — BFS-flatten embedded interfaces, copy methods от embeds к parent. Missing embedded interfaces оставляют parent с fewer methods.

### AST Канонизация

Пакет `ast_stringer/` (private) converts AST nodes to canonical Go syntax strings. Key behaviors:

- **Порядок разрешения идентификаторов**: import alias → generic alias → builtin type → package prefix
- **Сортировка методов интерфейса**: методы сортируются alphabetically для canonical representation (Go interface equality ignores method order)
- **Composite типы**: `chan<- Event`, `map[string][]int`, `func(string) error` — все сериализуются в deterministic string form

## Этап 3: Разрешение зависимостей

### Двухиндексная архитектура

Ресолвер строит два complementary индекса из `ParsedData` (`build_index.go`):

#### Type Index (`byType`)

```go
map[string][]*domain.Provider  // normalized result type → candidate providers
```

Строит ключ вызовом `normalizeGenericParameter()` на `Result` каждого провайдера:
- `RedisStorage[User, Order]` → `RedisStorage[T, T]`
- `RedisStorage[T, T1]` → `RedisStorage[T, T]`
- `Cache[string, int]` → `Cache[T, T]`

Каждый провайдер индексируется под своим normalized key. Non-generic провайдеры также индексируются по exact result.

#### Interface Method Index (`byMethod`)

```go
map[string][]*domain.Provider  // method signature → providers implementing it
```

Для result типа каждого провайдера, если он имплементирует интерфейс, каждый метод индексируется через `methodKey()`:

| Тип метода | Пример ключа индекса |
|------------|---------------------|
| Non-generic | `Get(id int) (*User, error)` |
| Generic | `Get` (bare name only) |

При разрешении queried both full и bare keys.

### Разрешение фасадов

Интерфейс qualifies как DI-фасад когда его `InterfaceID` starts with `module/DIDir/` (`resolve_facades.go`). Например, с модулем `github.com/diplexhq/diplex` и `-di internal/di`, любой интерфейс на `github.com/diplexhq/diplex/internal/di/...` becomes a facade. Result type каждого facade method seeds the BFS queue.

### BFS разрешение провайдеров

Разрешение провайдеров использует stack-based BFS (`resolve_providers.go`):

1. Извлечь все unique facade method result types в `queue`
2. Pop from stack, вызвать `findProviderCollection()`
3. Для каждого найденного провайдера добавить его argument types в queue (если не seen)
4. Повторять пока queue не пуст

```go
for len(queue) > 0 {
    arg := queue[len(queue)-1]
    queue = queue[:len(queue)-1]
    collection := findProviderCollection(arg, parsedData, index)
    providers[arg] = collection
    for _, p := range collection.Providers {
        for _, a := range p.Arguments {
            if !seen[a] { seen[a] = true; queue = append(queue, a) }
        }
    }
}
```

### Дипatch коллекций провайдеров

`findProviderCollection()` (`find_providers.go:14-38`) handles slice vs. single:

- Если arg starts with `[]` → strip prefix, recurse, wrap в `ProviderCollection{CollectionType: "slice", ...}`
- Nested collections panic
- Single providers sorted by ID для deterministic output

Dispatch к type или interface matching через `findProviders()`:

```go
if parsedData.Interfaces[arg].RealType == false && has methods → findProvidersByInterface()
else → findProvidersByType()
```

### Нормализация generic параметров

`normalizeGenericParameter()` (`utils.go:32-77`) заменяет все generic type аргументы внутри `[...]` на `"T"`. Handle nested generics, pointers, composite types:

```
pkg.Cache[order.Order, payment.Payment]  → pkg.Cache[T, T]
*pkg.Cache[T0, T1]                       → *pkg.Cache[T, T]
pkg.Cache[repo.Repo[T0], T1]             → pkg.Cache[T, T]
map[int]string                           → map[int]string  (map brackets skipped)
```

### Алгоритм type matching

`findProvidersByType()` (`find_providers_by_type.go`):

1. Normalize wanted type через `normalizeGenericParameter()`
2. Lookup в `typeIndex`
3. Для каждого candidate, clone его generic constraints
4. Run `compareParams()` match provider result к wanted type
5. Если match, вызвать `resolveProvider()` для генерации concrete instantiations

#### compareParams — Character-by-Character Token Matching

`compareParams()` (`find_providers.go:60-90`) проверяет если provider result type satisfies wanted parameter type:

1. **Exact match** — strings identical → `true`
2. **Empty constraints** — нет generic constraints и strings differ → `false`
3. **Token loop** — consume оба strings char-by-char:
   - Получить next token из обоих strings
   - Если provider token **non-generic** (не `T`, `T0`, etc.) → exact match required
   - Если provider token **generic** (`T`/`T0`/`T1`) → extract concrete type из wanted string и `squeezeConstraint()`
4. **Оба consumed** — после loop оба strings должны быть fully consumed

#### squeezeConstraint — Generic Constraint Narrowing

`squeezeConstraint()` (`find_providers.go:95-107`) — core generic resolution:

- Если paramName уже в constraints → verify concrete type в allowed list (или `"any"`), затем **narrow** list до single concrete type
- Если paramName new → create entry с concrete type

Это **progressive narrowing** процесс. Каждый `compareParams()` вызов по BFS path further restricts constraints map. Когда все методы matched, constraints содержат concrete types для instantiation.

```go
func squeezeConstraint(constraints, paramName, concreteType) bool {
    if existing, ok := constraints[paramName]; ok {
        if !Contains(existing, concreteType) && !Contains(existing, "any") {
            return false  // constraint violation
        }
        constraints[paramName] = []string{concreteType}  // narrow to single
    } else {
        constraints[paramName] = []string{concreteType}  // new
    }
    return true
}
```

### Алгоритм interface matching

`findProvidersByInterface()` (`find_providers_by_interface.go`):

1. Lookup both full method key и bare name key в `byMethod` index
2. Для каждого candidate:
   - candidate result type == wanted → direct match
   - candidate generic → extract prototype, clone constraints, narrow via `methodMatches()`
3. `methodMatches()` итерирует все wanted methods, verifying:
   - Provider has each method
   - Argument count matches
   - Result count matches
   - Each argument/result passes `compareParams()` с current constraints

#### Constraint propagation для интерфейсов

При matching generic providers к interfaces:

```go
// Step 1: Clone provider original constraints (от type param declaration)
providerConstraints := maps.Clone(candidate.Generic)

// Step 2: Narrow через method matching results
for k, v := range constraints {  // constraints от instantiate generic
    if originalParam, ok := providerConstraints[v[0]]; ok {
        providerConstraints[originalParam] = narrowed[k]
    }
}

// Step 3: Generate all combinations и создать concrete providers
resolveProvider(candidate, providerConstraints)
```

### Комбинаторное generic resolution

Когда generic provider constraints содержат multiple options (e.g., `K: ["string", "int"]`, `V: ["User", "Order"]`), `generateCombinations()` (`find_providers.go:150-187`) производит Cartesian product:

```
K ∈ {string, int}
V ∈ {User, Order}

Combinations:
  1. {K: string, V: User}
  2. {K: string, V: Order}
  3. {K: int, V: User}
  4. {K: int, V: Order}
```

Combination generator использует recursive DFS over **sorted** constraint keys для determinism:

```
generate(idx, current):
    if idx == len(keys): result.push(clone(current)); return
    k := keys[idx]
    for each option in constraints[k]:
        current[k] = option
        generate(idx+1, current)
```

Каждая combination produces concrete `*domain.Provider` через `utils.ReplaceTokens()`, substituting generic names с concrete types в Result, Arguments, Name, ID.

### Disambiguation по имени

Когда multiple providers match single parameter type (`len(collection.Providers) > 1`), `selectSingleProvider()` (`resolve.go:95-117`) resolves ambiguity:

1. Collection имеет ровно 1 provider → return it
2. `argName` empty → return nil (caller panics)
3. Match `provider.ResultName` against `argName` через `matchByName()`:
   - Equal length
   - First character case-insensitive match
   - Rest exact match
4. Exactly 1 match → return it; иначе → nil (caller panics)

`ResultName` derivation: first named return value, или local name of result type (`*UserRepo` → `UserRepo` → `userrepo` lowercase).

## Этап 4: Генерация кода

### Template system

Четыре embedded templates combined per DI facade directory (`internal/generator/tmpl/`):

| Template | Purpose |
|----------|---------|
| `head.tmpl` | Package declaration и file comment |
| `import.tmpl` | Grouped imports с unique aliases |
| `facade.tmpl` | Facade method implementations delegating к DI struct |
| `provider.tmpl` | DI struct с `func()` fields + `sync.OnceValue` wrappers |

### Dependency data construction

`buildData()` делает **BFS traversal** dependency graph от facade methods:

1. Для каждого facade method, identify resolved provider
2. Recursively traverse `ArgumentProviders` для каждого provider
3. Track unique packages, providers, methods
4. Sort all collections deterministically (providers by ID, methods alphabetically)

### Package alias resolution

Imports grouped и sorted:

1. **Standard library** — no alias если base name matches package path suffix
2. **Third-party** — alphabetical, с `idUniq()` для collision avoidance
3. **Local packages** — после module prefix stripped, тот же dedup logic

### Provider field generation

Каждое provider поле в DI struct — `func() ResultType` wrapped в `sync.OnceValue()`:

```go
di.entityNewUserRepo = sync.OnceValue(func() *entity.UserRepo {
    return entity.NewUserRepo(
        di.storageNewRedisStorageEntityUserString(),
        di.storageNewStringEntityUser(),
    )
})
```

- Providers returning `error` include: `result, err := fn(); if err != nil { panic(err.Error()) }; return result`
- Slice providers return `[]T{di.provider1(), di.provider2(), ...}`
- Field names sanitized к camelCase: `entity.NewRepo[entity.Order]` → `entityNewRepoEntityOrder`

### Determinism guarantees

- Все map iterations заменены sorted key traversal
- Generic constraint keys sorted before combinatorial generation
- Provider collections sorted by ID before template rendering
- Import aliases use `idUniq()` для collision-free naming
- Result: identical source → identical `di.go` output (verified by SHA-256 hash)

## Data Flow Summary

```
Scanner.Scan()
  │  single goroutine, filepath.WalkDir, skipPattern filters
  ▼
chan SourceFile  (buffered, capacity 4)
  │
  ├─ 4 parser goroutines (sync.WaitGroup)
  ├─ parseState protected by sync.Mutex
  ├─ Post-parse: resolveAliases() + resolveEmbeds()
  ▼
ParsedData { Providers, Interfaces }
  │
  ├─ buildTypeIndex() → normalizeGenericParameter()
  ├─ buildInterfaceIndex() → methodKey()
  ├─ resolveFacades() → DIDir prefix match
  ├─ BFS queue: facade results → provider args → ...
  ├─ compareParams() → squeezeConstraint() → progressive narrowing
  ├─ generateCombinations() → Cartesian product of constraints
  ├─ selectSingleProvider() → matchByName() disambiguation
  ▼
ResolvedData { ResolvedFacades, ResolvedProviders }
  │
  ├─ buildData() → BFS dep tree, sorted imports, sorted providers
  ├─ 4 templates: head + import + facade + provider
  ├─ sync.OnceValue() lazy init wrappers
  └─ SHA-256 deterministic verification
  ▼
di.go (output)
```

## Performance Characteristics

| Stage | Complexity | Notes |
|-------|-----------|-------|
| Scanner | O(n) где n = files in tree | Single goroutine, buffered channel |
| Parser | O(m × f) где m = files, f = AST nodes | 4 workers, mutex-protected state |
| Index building | O(p × k) где p = providers, k = methods | Linear scan of parsed data |
| BFS resolution | O(d × c) где d = dependency depth, c = candidates | Short-circuits on exact match |
| Constraint narrowing | O(t × n) где t = type tokens, n = constraints | Character-by-character scan |
| Combinatorial gen | O(∏cᵢ) где cᵢ = constraint cardinality | Bounded by constraint size |
| Generation | O(p × t) где p = providers, t = templates | Sorted iterations, deterministic |

## Error Boundaries

| Scenario | Error mechanism | Recovery |
|----------|----------------|----------|
| No provider found | `panic("cannot generate DI: no providers for ...")` | None — build fails |
| Ambiguous match | `panic("no single provider found for ...")` | None — rename parameter |
| Cycle detection | None (runtime panic via sync.OnceValue) | Redesign dependency graph |
| Invalid AST | `astStringer` returns empty string | Skip entity |
| Alias chain cycle | `utils.ResolveReplacements()` panics | Fix alias declaration |
