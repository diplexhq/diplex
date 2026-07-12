# Краткая справка (Cheat Sheet)

## Структура проекта (быстрый обзор)

```
diplex/
├── main.go              # оркестратор
├── doc/                 # документация
└── internal/
    ├── scanner/         # поиск .go файлов (chan SourceFile)
    ├── parser/          # парсинг и анализ AST
    │   └── ast_stringer/ # AST → string (приватный)
    ├── domain/          # доменные типы (ParsedData, Provider, MethodContract)
    ├── generator/       # генерация DI-кода
    │   └── tmpl/        # шаблоны
    └── utils/           # Must/NoErr/Logger/SanitizeIdent
```

## Разрешение зависимостей (Resolver)

Ресолвер строит два индекса из `ParsedData` и выполняет BFS-обход от facade-интерфейсов:

1. **`typeIndex`** — маппит нормализованные Result-типы провайдеров на кандидатов. Generic-типы нормализуются: все параметры `T`, `T0`, `T1` → `T`.
2. **`interfaceMethodIndex`** — маппит сигнатуры методов на провайдеры, которые их реализуют. Non-generic методы: `"Name(params)(results)"`, generic: bare имя метода.

**Процесс разрешения:**
- `resolveFacades()` — фильтрация DI facade интерфейсов по `-di` директориям
- `resolveProviders()` — BFS обход от return-типов методов facades, сбор всех транзитивных зависимостей
- `findProviderCollection()` — поиск по типу (typeIndex) или по интерфейсу (interfaceMethodIndex)
- Generic constraints сужаются через `compareType()`, декартово произведение для multi-constraint

**Результат:** `ResolvedData{Facades: Interfaces, Providers: map[Parameter]ProviderCollection}`

## Ключевые типы

### SourceFiles канал

```go
// domain.SourceFiles — канал путей к исходным файлам
type SourceFiles chan SourceFile
```

Scanner создаёт канал в фоновой горутине и закрывает его после заполнения. Parser читает канал 4 воркерами.

### domain.ParsedData

```go
// domain.ParsedData — распарсенный граф зависимостей
type ParsedData struct {
    Providers  Providers       // функции New* (конструкторы)
    Interfaces Interfaces      // интерфейсы + методы структур (RealType: true)
}

// domain.Provider — контракт конструктора с поддержкой дженериков
type Provider struct {
    Pkg       string            // путь пакета
    Name      string            // имя функции (New*)
    Arguments []Parameter       // строки типов параметров
    Result    Parameter         // строка типа результата
    Generic   map[string][]string // T → [тип constraints]
    Error     bool              // возвращает error
}
// Provider.Id() возвращает синтетический ID: pkg + "." + name

// domain.MethodMap — имя метода → контракт
type MethodMap map[FunctionName]MethodContract

// domain.MethodContract — сигнатура одного метода
type MethodContract struct {
    Arguments []Parameter
    Results   []Parameter
}
```

## Как запустить

```bash
# Запуск из корня целевого проекта
go tool diplex

# Пользовательские директории сканирования
go tool diplex -scan internal,pkg

# Пользовательский вывод
go tool diplex -out custom/dir

# Пропустить go.mod, задать модуль вручную
go tool diplex -module example.com/my/project

# Пользовательский паттерн пропуска
go tool diplex -skip "(generated|mocks|_test\.go)"

# Подробный / тихий режим
go tool diplex -v
go tool diplex -s
```

## Сгенерированные файлы

`internal/generated/diplex/`
- `di.go` — DI-фасад (автоматически сгенерированный, реализует `di.DI` интерфейс)

DI-фасад генерируется для каждой директории из флага `-di`. Каждая директория производит один `.go` файл со структурой `DI` и её методами, которые разрешают все зависимости через `sync.OnceValue`.

### Пример использования

```go
// Ваш исходный код (сканируется diplex):

// 1. Интерфейс
type UserService interface {
    GetUser(id string) (*User, error)
}

// 2. Реализация
type service struct { repo *UserRepo }
func (s *service) GetUser(id string) (*User, error) { /* ... */ }

// 3. Конструктор
func NewService(repo *UserRepo) (*service, error) {
    return &service{repo: repo}, nil
}

// После запуска diplex, используйте сгенерированный контейнер:

func main() {
    // 1. Создайте контейнер (сгенерирован diplex)
    c := diplex.NewDI()

    // 2. Получите сервисы — зависимости разрешены автоматически
    svc := c.UserService()

    // 3. Используйте
    user, err := svc.GetUser("user-123")
}
```

## Паттерны распознавания

**Провайдер:**
```go
func NewX(deps...) (*X, error?)  // ✓
func NewRepo[T User | Order]() *Repo[T]  // ✓ generic → concrete instantiation
```

**Интерфейс:**
```go
type Interface interface {       // ✓
    Method(args) results
}
type Repository[T any] interface { // ✓ generic interface
    Get(id int) (T, error)
}
```

**Реализация:**
```go
func (x *X) Method(args) results // ✓
func (r *Repo[Order]) Get(id int) (Order, error) // ✓ generic receiver
```

**Коллекция (slice):**
```go
func NewServer(handlers []Handler) *Server // ✓ агрегирует всех Handler-провайдеров
```

## Важные файлы

| Файл | Назначение |
|------|------------|
| [ru/PROJECT.md](ru/PROJECT.md) | Полная архитектура |
| [ru/REQUIREMENTS.md](ru/REQUIREMENTS.md) | Требования |
| [ru/DEVELOPMENT.md](ru/DEVELOPMENT.md) | Руководство разработчика |
| [ru/QUICK_REF.md](ru/QUICK_REF.md) | Эта шпаргалка |
| [ru/TESTING.md](ru/TESTING.md) | Правила тестирования |
