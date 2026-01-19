# Postgres Migration Utility

A PostgreSQL migration utility with automatic synchronization of database state and file system.

![Tests](https://github.com/quadgod/pgm/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/quadgod/pgm/branch/main/graph/badge.svg)](https://codecov.io/gh/quadgod/pgm)

**Read this in other languages:** [Русский](README.ru.md)

## Project Purpose

This migrator solves the problem of synchronization between database state on dev environments and migrations in the repository. When multiple developers work on a project simultaneously and apply migrations to shared dev environments, conflicts and desynchronization often occur.

### Key Feature

The utility can:
- **Compare** database state with migrations in the file system
- **Automatically rollback** database changes until it fully matches the state of migrations in the folder (when using `--priority=fs`)
- **Synchronize** dev environments with the current state of code in the repository
- **Safely apply** only new migrations in production (when using `--priority=db`)

The behavior is controlled by the `--priority` parameter:
- `--priority=fs`: Development mode - automatically synchronizes database with file system (can rollback migrations)
- `--priority=db`: Production mode - safely applies only missing migrations, fails on conflicts

See the [Running the Migrator](#running-the-migrator) section for detailed documentation.

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

## Working with Taskfile

The project uses [Taskfile](https://taskfile.dev/) for command management. The root `taskfile.yml` includes nested taskfiles from other project directories via the `includes` mechanism.

### Taskfile Structure in the Project

- **Root** `taskfile.yml` - main project commands (lint, test)
- **`cmd/pgm/taskfile.yml`** - commands for building pgm (prefix: `pgm:`)
- **`sandbox/migrations/taskfile.yml`** - commands for working with sandbox migrations (prefix: `migsbox:`)

### Running Commands from Nested Taskfiles

When you're in the project root directory, you can run commands from nested taskfiles using prefixes defined in the root taskfile's `includes` section:

#### Commands for building pgm

```bash
# From project root
task pgm:build          # Builds pgm for all platforms
```

#### Commands for sandbox migrations

```bash
# From project root
task migsbox:pgstart    # Starts PostgreSQL in Docker for sandbox
task migsbox:pgstop     # Stops PostgreSQL in Docker
task migsbox:create -- "migration_name"  # Creates a new migration
task migsbox:migrate    # Applies all migrations
task migsbox:down       # Rolls back the last migration
```

#### Running commands from current directory

If you're inside a directory with a taskfile (e.g., in `sandbox/migrations/`), you can run commands without the prefix:

```bash
# When inside sandbox/migrations/
cd sandbox/migrations/
task create -- "migration_name"  # Creates a migration
task migrate                      # Applies migrations
task down                         # Rolls back migration
task pgstart                      # Starts PostgreSQL
```

#### Referencing root taskfile commands

If a nested taskfile needs to call a command from the root taskfile, use the `:` prefix (colon at the beginning):

```yaml
# In sandbox/migrations/taskfile.yml
deps:
  - task: :pgm:build  # Calls task pgm:build from root taskfile
```

### Available Commands List

**Root commands:**
- `task lint` - code linting
- `task test` - run tests
- `task test:cover` - run tests with coverage

**Build commands (prefix `pgm:`):**
- `task pgm:build` - build pgm for all platforms

**Sandbox commands (prefix `migsbox:`):**
- `task migsbox:pgstart` - start PostgreSQL in Docker
- `task migsbox:pgstop` - stop PostgreSQL in Docker
- `task migsbox:create -- "name"` - create a new migration
- `task migsbox:migrate` - apply all migrations
- `task migsbox:down` - roll back the last migration

### Additional Information

To view all available commands, use:

```bash
task --list        # Shows all root taskfile commands
task --list-all    # Shows all commands including nested ones
```

For more information about Taskfile: https://taskfile.dev/

## Running the Migrator

The `pgm` utility provides a command-line interface for managing PostgreSQL migrations. Below is a detailed description of how to use it and its parameters.

### Installation

Install `pgm` using Go:

```bash
go install github.com/quadgod/pgm/cmd/pgm@latest
```

### CLI Parameters

The migrator accepts the following command-line parameters:

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `--command` | Yes | - | Command to execute: `create`, `migrate`, or `down` |
| `--migrationsDir` | Yes | - | Path to the directory containing migration files |
| `--migrationName` | For `create` | - | Name of the migration (alphanumeric and `_` only) |
| `--migrationsTableSchema` | For `migrate`/`down` | - | Schema name for the migrations table |
| `--migrationsTable` | For `migrate`/`down` | `migrations` | Name of the migrations table |
| `--connectionString` | For `migrate`/`down` | `PG_CONNECTION_STRING` env var | PostgreSQL connection string |
| `--priority` | No | `fs` | Priority mode: `fs` (file system) or `db` (database) |

### Commands

#### Create Migration

Creates a new migration file pair (`{timestamp}_{name}.up.sql` and `{timestamp}_{name}.down.sql`):

```bash
pgm \
  --command=create \
  --migrationsDir=./migrations \
  --migrationName=add_users_table
```

#### Migrate (Apply Migrations)

Applies migrations to the database. Behavior depends on the `--priority` parameter (see below).

```bash
pgm \
  --command=migrate \
  --migrationsDir=./migrations \
  --migrationsTableSchema=public \
  --migrationsTable=migrations \
  --connectionString="postgres://user:password@localhost:5432/dbname" \
  --priority=fs
```

#### Down (Revert Migration)

Reverts the last applied migration:

```bash
pgm \
  --command=down \
  --migrationsDir=./migrations \
  --migrationsTableSchema=public \
  --migrationsTable=migrations \
  --connectionString="postgres://user:password@localhost:5432/dbname" \
  --priority=fs
```

### Priority Parameter: Key Feature

The `--priority` parameter is a **critical feature** that determines how the migrator handles conflicts between the database state and migration files. This parameter has two modes:

#### `--priority=fs` (File System Priority) - Development Mode

**Use this mode for development environments.**

When `priority=fs` is set, the migrator:
- **Prioritizes migration files** in the file system over the database state
- **Automatically rolls back** migrations that exist in the database but don't match the file system
- **Synchronizes** the database to match the exact state of migration files in the repository
- **Applies missing migrations** from the file system after rollback

**Behavior:**
1. Compares migrations in the database with migration files
2. If a mismatch is found, automatically reverts all conflicting migrations from the database
3. Applies migrations from the file system to bring the database to the expected state

**Example scenario:**
- Database has migrations: `001_init`, `002_add_users`, `003_add_posts`
- File system has migrations: `001_init`, `002_add_users`
- Result: Migration `003_add_posts` is automatically reverted, database matches file system

**⚠️ Warning:** This mode can cause data loss if migrations are reverted. Use only in development environments.

#### `--priority=db` (Database Priority) - Production Mode

**Use this mode for production environments.**

When `priority=db` is set, the migrator:
- **Prioritizes the database state** over migration files
- **Only applies missing migrations** that exist in the file system but not in the database
- **Returns an error** if migrations in the database don't match the file system
- **Never automatically rolls back** migrations

**Behavior:**
1. Compares migrations in the database with migration files
2. Applies only migrations that are missing from the database
3. If a mismatch is found (different migration name at the same position), returns an error and stops

**Example scenario:**
- Database has migrations: `001_init`, `002_add_users`
- File system has migrations: `001_init`, `002_add_users`, `003_add_posts`
- Result: Migration `003_add_posts` is applied
- If database had `002_different_migration` instead of `002_add_users`, an error would be returned

**✅ Safe for production:** This mode never modifies existing migrations and only adds new ones.

### Summary: When to Use Which Priority

| Environment | Priority | Reason |
|-------------|----------|--------|
| Development/Staging | `fs` | Allows automatic synchronization and rollback for testing |
| Production | `db` | Safe mode that only applies new migrations, prevents accidental rollbacks |
| CI/CD pipelines | `db` | Predictable behavior, fails fast on conflicts |
| Local development | `fs` | Convenient synchronization with repository state |

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
