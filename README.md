# Postgres Migration Utility

## Структура репозитория

- [Документация golang best practices для орагизации структуры репозитория](https://github.com/golang-standards/project-layout/blob/master/README_ru.md)

## Зависимости

- brew install go (min version golang 1.25.5)
- brew install golangci-lint
- Управление командами https://taskfile.dev/
  - install on macos `brew install go-task`
  - install on windows `choco install go-task`
  - install on linux
    - aur `yay -S go-task`
    - fedora `sudo dnf install go-task`

## Состав репоизитория

- [pgm](./cmd/pgm/README.md) - утилита для миграции postgres

## Команды

- `task lint` - запускает проверку кода
- `task test` - запускает тесты
- `task pgm:build` - собирает pgm
