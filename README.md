# Postgres Migration Utility

A PostgreSQL migration utility with automatic synchronization of database state and file system.

**Read this in other languages:** [Русский](README.ru.md)

## Project Purpose

This migrator solves the problem of synchronization between database state on dev environments and migrations in the repository. When multiple developers work on a project simultaneously and apply migrations to shared dev environments, conflicts and desynchronization often occur.

### Key Feature

The utility can:
- **Compare** database state with migrations in the file system
- **Automatically rollback** database changes until it fully matches the state of migrations in the folder
- **Synchronize** dev environments with the current state of code in the repository

### Problem Being Solved

On dev environments, the following situations often arise:
- Developer A applies their migration directly to the database
- Developer B applies a different migration to the same database
- Migrations conflict or create incompatible changes
- On the next deploy or code update, an error occurs because migrations are not applied in the correct order or the database state does not match the expected state

**Solution**: `pgm` automatically brings the database to a state that matches the migration files in the repository, rolling back extra changes and applying missing migrations in the correct order.

### Benefits

- ✅ Dev environments always match the code in the repository
- ✅ No need to manually resolve migration conflicts
- ✅ Predictable database state before each deploy
- ✅ Simplified team workflow with shared environments

## Repository Structure

- [Golang best practices documentation for repository organization](https://github.com/golang-standards/project-layout)

## Dependencies

- `brew install go` (min version golang 1.25.5)
- `brew install golangci-lint`
- Command management via [Taskfile](https://taskfile.dev/):
  - Install on macOS: `brew install go-task`
  - Install on Windows: `choco install go-task`
  - Install on Linux:
    - AUR: `yay -S go-task`
    - Fedora: `sudo dnf install go-task`

## Repository Contents

- [pgm](./cmd/pgm/README.md) - PostgreSQL migration utility

## Commands

- `task lint` - runs code linting
- `task test` - runs tests
- `task pgm:build` - builds pgm

## Example: How to Use in Other Projects

```yml
# https://taskfile.dev

version: "3"

vars:
  BIN_DIR: "{{.ROOT_DIR}}/bin"
  EXE: '{{if eq OS "windows"}}.exe{{end}}'

  PGM_PKG: "github.com/quadgod/pgm/cmd/pgm"
  PGM_VER: "latest"
  PGM_BIN: "{{.BIN_DIR}}/pgm{{.EXE}}"

  # Where migrations are located in your project (change to match your setup)
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

### Linux & macOS Example

```bash
# bash
DB_DSN='postgres://u:p@host:5432/db' task migrate:up
```

### Windows Example

```powershell
# powershell
$env:DB_DSN="postgres://u:p@host:5432/db" task migrate:up
```
