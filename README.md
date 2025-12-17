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

## Example how to use in other project

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
    desc: "Run migrations up"
    deps: [tools:pgm]
    cmds:
      - cmd: >
          "{{.PGM_BIN}}"
          --command=migrate
          --migrationsDir="{{.MIGRATIONS_DIR}}"
          --migrationsTableSchema="{{.MIGRATIONS_SCHEMA}}"
          --migrationsTable="{{.MIGRATIONS_TABLE}}"
          --connectionString="{{.DB_DSN}}"

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
```

### Linux & MacOS example

```bash
# bash
DB_DSN='postgres://u:p@host:5432/db' task migrate:up
```

### Windows example

```powershell
# powershell
$env:DB_DSN="postgres://u:p@host:5432/db" task migrate:up
```
