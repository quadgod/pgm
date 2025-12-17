package pgm

import (
	"fmt"
)

type Priority string

const (
	DB Priority = "db"
	FS Priority = "fs"
)

type Command string

const (
	CREATE  Command = "create"
	MIGRATE Command = "migrate"
	DOWN    Command = "down"
)

type MigratorOptions struct {
	Priority              Priority
	Command               Command
	MigrationName         string
	MigrationsDir         string
	MigrationsTableSchema string
	MigrationsTable       string
	ConnectionString      string
}

func (o *MigratorOptions) MigrationsTableNameWithSchema() string {
	return fmt.Sprintf("%s.%s", o.MigrationsTableSchema, o.MigrationsTable)
}
