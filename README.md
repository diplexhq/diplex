# Gendior — генератор DI-контейнера для Go

Утилита для сканирования Go-кода и генерации высокоэффективного DI-контейнера на этапе компиляции. Без runtime reflection — сгенерированный код не уступает ручному внедрению зависимостей.

---

# Gendior — Go DI Container Generator

Utility for scanning Go code and generating a high-performance DI container at compile time. No runtime reflection — generated code equals manual dependency injection.

## Quick Start / Быстрый старт

```bash
# Из корня целевого проекта / From target project root
go tool gendior
```

## Documentation / Документация

| | English | Русский |
|---|---------|---------|
| Project Architecture | [doc/en/PROJECT.md](doc/en/PROJECT.md) | [doc/ru/PROJECT.md](doc/ru/PROJECT.md) |
| Requirements | [doc/en/REQUIREMENTS.md](doc/en/REQUIREMENTS.md) | [doc/ru/REQUIREMENTS.md](doc/ru/REQUIREMENTS.md) |
| Quick Reference | [doc/en/QUICK_REF.md](doc/en/QUICK_REF.md) | [doc/ru/QUICK_REF.md](doc/ru/QUICK_REF.md) |
| TODO / Planned | [doc/en/TODO.md](doc/en/TODO.md) | [doc/ru/TODO.md](doc/ru/TODO.md) |

## Key Features / Ключевые возможности

- **Compile-time DI** — без runtime reflection, максимальная производительность
- **Type-safe** — все зависимости проверяются при компиляции
- **Zero overhead** — сгенерированный код равен ручному внедрению
- **Standard library only** — только стандартная библиотека Go

## Limitations / Ограничения

- Только публичные типы (с заглавной буквы) / Only public types (capitalized)
- Один конструктор на тип / One constructor per type
- Без поддержки дженериков / No generics support
