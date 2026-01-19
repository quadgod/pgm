# Postgres Migration Utility

Утилита для управления миграциями PostgreSQL с автоматической синхронизацией состояния базы данных и файловой системы.

![Tests](https://github.com/quadgod/pgm/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/quadgod/pgm/branch/main/graph/badge.svg)](https://codecov.io/gh/quadgod/pgm)

**Read this in other languages:** [English](README.md)

## Назначение проекта

Этот мигратор решает проблему рассинхронизации между состоянием базы данных на dev-стендах и миграциями в репозитории. Когда несколько разработчиков одновременно работают над проектом и вносят миграции на общие dev-стенды, часто возникают конфликты и рассинхронизация.

### Ключевая особенность

Утилита умеет:
- **Сравнивать** состояние базы данных с миграциями в файловой системе
- **Автоматически откатывать** изменения в базе данных до момента, когда она полностью соответствует состоянию миграций в папке (при использовании `--priority=fs`)
- **Синхронизировать** dev-стенды с актуальным состоянием кода в репозитории
- **Безопасно применять** только новые миграции в продакшене (при использовании `--priority=db`)

Поведение контролируется параметром `--priority`:
- `--priority=fs`: Режим разработки - автоматически синхронизирует базу данных с файловой системой (может откатывать миграции)
- `--priority=db`: Режим продакшена - безопасно применяет только недостающие миграции, останавливается при конфликтах

Подробную документацию см. в разделе [Запуск мигратора](#запуск-мигратора).

### Решаемая проблема

На dev-стендах часто возникают ситуации:
- Разработчик A применяет свою миграцию напрямую на БД
- Разработчик B применяет другую миграцию на ту же БД
- Миграции конфликтуют или создают несовместимые изменения
- При следующем деплое или обновлении кода возникает ошибка, так как миграции не применяются в правильном порядке или состояние БД не соответствует ожидаемому

**Решение**: `pgm` автоматически приводит базу данных в состояние, соответствующее файлам миграций в репозитории, откатывая лишние изменения и применяя недостающие миграции в правильном порядке.

### Преимущества

- ✅ Dev-стенды всегда соответствуют коду в репозитории
- ✅ Нет необходимости вручную разбираться с конфликтами миграций
- ✅ Предсказуемое состояние базы данных перед каждым деплоем
- ✅ Упрощенная работа команды разработки с общими стендами

## Структура репозитория

- [Документация golang best practices для организации структуры репозитория](https://github.com/golang-standards/project-layout/blob/master/README_ru.md)

## Зависимости

- `brew install go` (min version golang 1.25.5)
- `brew install golangci-lint`
- Управление командами через [Taskfile](https://taskfile.dev/):
  - install on macos `brew install go-task`
  - install on windows `choco install go-task`
  - install on linux
    - aur `yay -S go-task`
    - fedora `sudo dnf install go-task`

## Состав репозитория

- [pgm](./cmd/pgm/README.md) - утилита для миграции postgres

## Команды

- `task lint` - запускает проверку кода
- `task test` - запускает тесты
- `task pgm:build` - собирает pgm

## Запуск мигратора

Утилита `pgm` предоставляет интерфейс командной строки для управления миграциями PostgreSQL. Ниже приведено подробное описание использования и параметров.

### Установка

Установите `pgm` с помощью Go:

```bash
go install github.com/quadgod/pgm/cmd/pgm@latest
```

### Параметры CLI

Мигратор принимает следующие параметры командной строки:

| Параметр | Обязательный | По умолчанию | Описание |
|----------|--------------|--------------|----------|
| `--command` | Да | - | Команда для выполнения: `create`, `migrate` или `down` |
| `--migrationsDir` | Да | - | Путь к директории с файлами миграций |
| `--migrationName` | Для `create` | - | Имя миграции (только буквы, цифры и `_`) |
| `--migrationsTableSchema` | Для `migrate`/`down` | - | Имя схемы для таблицы миграций |
| `--migrationsTable` | Для `migrate`/`down` | `migrations` | Имя таблицы миграций |
| `--connectionString` | Для `migrate`/`down` | Переменная `PG_CONNECTION_STRING` | Строка подключения к PostgreSQL |
| `--priority` | Нет | `fs` | Режим приоритета: `fs` (файловая система) или `db` (база данных) |

### Команды

#### Создание миграции

Создает новую пару файлов миграции (`{timestamp}_{name}.up.sql` и `{timestamp}_{name}.down.sql`):

```bash
pgm \
  --command=create \
  --migrationsDir=./migrations \
  --migrationName=add_users_table
```

#### Применение миграций (Migrate)

Применяет миграции к базе данных. Поведение зависит от параметра `--priority` (см. ниже).

```bash
pgm \
  --command=migrate \
  --migrationsDir=./migrations \
  --migrationsTableSchema=public \
  --migrationsTable=migrations \
  --connectionString="postgres://user:password@localhost:5432/dbname" \
  --priority=fs
```

#### Откат миграции (Down)

Откатывает последнюю примененную миграцию:

```bash
pgm \
  --command=down \
  --migrationsDir=./migrations \
  --migrationsTableSchema=public \
  --migrationsTable=migrations \
  --connectionString="postgres://user:password@localhost:5432/dbname" \
  --priority=fs
```

### Параметр Priority: Ключевая особенность

Параметр `--priority` — это **ключевая особенность**, которая определяет, как мигратор обрабатывает конфликты между состоянием базы данных и файлами миграций. Этот параметр имеет два режима:

#### `--priority=fs` (Приоритет файловой системы) - Режим разработки

**Используйте этот режим для dev-стендов.**

Когда установлен `priority=fs`, мигратор:
- **Приоритизирует файлы миграций** в файловой системе над состоянием базы данных
- **Автоматически откатывает** миграции, которые существуют в базе данных, но не соответствуют файловой системе
- **Синхронизирует** базу данных, чтобы она соответствовала точному состоянию файлов миграций в репозитории
- **Применяет недостающие миграции** из файловой системы после отката

**Поведение:**
1. Сравнивает миграции в базе данных с файлами миграций
2. Если обнаружено несоответствие, автоматически откатывает все конфликтующие миграции из базы данных
3. Применяет миграции из файловой системы, чтобы привести базу данных к ожидаемому состоянию

**Пример сценария:**
- В базе данных есть миграции: `001_init`, `002_add_users`, `003_add_posts`
- В файловой системе есть миграции: `001_init`, `002_add_users`
- Результат: Миграция `003_add_posts` автоматически откатывается, база данных соответствует файловой системе

**⚠️ Предупреждение:** Этот режим может привести к потере данных при откате миграций. Используйте только в dev-окружениях.

#### `--priority=db` (Приоритет базы данных) - Режим продакшена

**Используйте этот режим для продакшена.**

Когда установлен `priority=db`, мигратор:
- **Приоритизирует состояние базы данных** над файлами миграций
- **Применяет только недостающие миграции**, которые существуют в файловой системе, но отсутствуют в базе данных
- **Возвращает ошибку**, если миграции в базе данных не соответствуют файловой системе
- **Никогда не откатывает миграции автоматически**

**Поведение:**
1. Сравнивает миграции в базе данных с файлами миграций
2. Применяет только миграции, которые отсутствуют в базе данных
3. Если обнаружено несоответствие (разное имя миграции на той же позиции), возвращает ошибку и останавливается

**Пример сценария:**
- В базе данных есть миграции: `001_init`, `002_add_users`
- В файловой системе есть миграции: `001_init`, `002_add_users`, `003_add_posts`
- Результат: Миграция `003_add_posts` применяется
- Если бы в базе данных была `002_different_migration` вместо `002_add_users`, была бы возвращена ошибка

**✅ Безопасно для продакшена:** Этот режим никогда не изменяет существующие миграции и только добавляет новые.

### Резюме: Когда использовать какой приоритет

| Окружение | Priority | Причина |
|-----------|----------|---------|
| Разработка/Staging | `fs` | Позволяет автоматическую синхронизацию и откат для тестирования |
| Продакшен | `db` | Безопасный режим, который применяет только новые миграции, предотвращает случайные откаты |
| CI/CD пайплайны | `db` | Предсказуемое поведение, быстрое обнаружение конфликтов |
| Локальная разработка | `fs` | Удобная синхронизация с состоянием репозитория |

## Пример: Как использовать в другом проекте

```yml
# https://taskfile.dev

version: "3"

vars:
  BIN_DIR: "{{.ROOT_DIR}}/bin"
  EXE: '{{if eq OS "windows"}}.exe{{end}}'

  PGM_PKG: "github.com/quadgod/pgm/cmd/pgm"
  PGM_VER: "latest"
  PGM_BIN: "{{.BIN_DIR}}/pgm{{.EXE}}"

  # где лежат миграции в проекте (поменяй под себя)
  MIGRATIONS_DIR: "{{.ROOT_DIR}}/migrations"
  MIGRATIONS_SCHEMA: "public"
  MIGRATIONS_TABLE: "migrations"
  DB_DSN:
    sh: 'echo "${DB_DSN:-postgres://sandbox:sandbox@localhost:5432/sandbox}"'
  GOLANGCI_LINT_PKG: "github.com/golangci/golangci-lint/cmd/golangci-lint"
  GOLANGCI_LINT_VER: "latest"
  GOLANGCI_LINT_BIN: "{{.BIN_DIR}}/golangci-lint{{.EXE}}"

tasks:
  tools:dir:
    desc: "Create local bin dir"
    cmds:
      - cmd: 'mkdir -p "{{.BIN_DIR}}"'
        platforms: [linux, darwin]
      - cmd: 'powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path ''{{.BIN_DIR}}'' | Out-Null"'
        platforms: [windows]

  tools:golangci-lint:
    desc: "Install golangci-lint into ./bin"
    deps: [tools:dir]
    cmds:
      - cmd: 'GOBIN="{{.BIN_DIR}}" go install {{.GOLANGCI_LINT_PKG}}@{{.GOLANGCI_LINT_VER}}'
        platforms: [linux, darwin]
      - cmd: 'powershell -NoProfile -Command "$env:GOBIN=''{{.BIN_DIR}}''; go install {{.GOLANGCI_LINT_PKG}}@{{.GOLANGCI_LINT_VER}}"'
        platforms: [windows]
    status:
      - '{{if eq OS "windows"}}powershell -NoProfile -Command "if (Test-Path ''{{.GOLANGCI_LINT_BIN}}'') { exit 0 } else { exit 1 }"{{else}}test -f "{{.GOLANGCI_LINT_BIN}}"{{end}}'

  lint:
    desc: "Run golangci-lint"
    deps: [tools:golangci-lint]
    cmds:
      - cmd: '"{{.GOLANGCI_LINT_BIN}}" run ./...'

  tools:pgm:
    desc: "Install pgm into ./bin"
    deps: [tools:dir]
    cmds:
      - cmd: 'GOBIN="{{.BIN_DIR}}" go install {{.PGM_PKG}}@{{.PGM_VER}}'
        platforms: [linux, darwin]
      - cmd: 'powershell -NoProfile -Command "$env:GOBIN=''{{.BIN_DIR}}''; go install {{.PGM_PKG}}@{{.PGM_VER}}"'
        platforms: [windows]
    status:
      - test -f "{{.PGM_BIN}}"

  tools:install:
    desc: "Install all tools"
    deps: [tools:pgm, tools:golangci-lint]

  migrate:create:
    desc: "Create migration. Usage: task migrate:create -- <migration_name>"
    deps: [tools:pgm]
    vars:
      MIGRATION_NAME: "{{.CLI_ARGS}}"
    requires:
      vars: [MIGRATION_NAME]
    cmds:
      - cmd: '"{{.PGM_BIN}}" --command=create --migrationsDir="{{.MIGRATIONS_DIR}}" --migrationName="{{.MIGRATION_NAME}}"'

  migrate:up:
    desc: "Run migrations up (dev mode: priority=fs)"
    deps: [tools:pgm]
    cmds:
      - cmd: >
          "{{.PGM_BIN}}"
          --command=migrate
          --migrationsDir="{{.MIGRATIONS_DIR}}"
          --migrationsTableSchema="{{.MIGRATIONS_SCHEMA}}"
          --migrationsTable="{{.MIGRATIONS_TABLE}}"
          --connectionString="{{.DB_DSN}}"
          --priority=fs

  migrate:up:prod:
    desc: "Run migrations up (production mode: priority=db)"
    deps: [tools:pgm]
    cmds:
      - cmd: >
          "{{.PGM_BIN}}"
          --command=migrate
          --migrationsDir="{{.MIGRATIONS_DIR}}"
          --migrationsTableSchema="{{.MIGRATIONS_SCHEMA}}"
          --migrationsTable="{{.MIGRATIONS_TABLE}}"
          --connectionString="{{.DB_DSN}}"
          --priority=db

  migrate:down:
    desc: "Run migration down"
    deps: [tools:pgm]
    cmds:
      - cmd: >
          "{{.PGM_BIN}}"
          --command=down
          --migrationsDir="{{.MIGRATIONS_DIR}}"
          --migrationsTableSchema="{{.MIGRATIONS_SCHEMA}}"
          --migrationsTable="{{.MIGRATIONS_TABLE}}"
          --connectionString="{{.DB_DSN}}"
          --priority=fs
```

### Пример для Linux & MacOS

```bash
# bash
DB_DSN='postgres://u:p@host:5432/db' task migrate:up
```

### Пример для Windows

```powershell
# powershell
$env:DB_DSN="postgres://u:p@host:5432/db" task migrate:up
```
