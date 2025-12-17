package pgm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Validate(t *testing.T) {
	t.Run("should return error if migrations dir is not set", func(t *testing.T) {
		flags := new(Flags)
		err := flags.Validate()

		assert.EqualError(t, err, "migrations dir is required")
	})

	t.Run("should not pass validation because invalid priority value", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = "invalid"
		flags.Command = "migrate"
		flags.MigrationsDir = "/migrations"
		flags.MigrationsTableSchema = "jobs"
		flags.MigrationsTable = "migrations"
		flags.ConnectionString = "some connection string"
		err := flags.Validate()

		assert.EqualError(t, err, "invalid priority. valid values \"db\" or \"fs\"")
	})

	t.Run("should return error if invalid command", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "unknown"
		flags.MigrationsDir = "/migrations"
		err := flags.Validate()

		assert.EqualError(t, err, "invalid command. valid cli \"create\", \"migrate\", \"down\"")
	})

	t.Run("should return error if create command & migration name contains forbidden symbols", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "create"
		flags.MigrationsDir = "/migrations"
		flags.MigrationName = "Hello_i contain_forbidden_symbol"
		err := flags.Validate()

		assert.EqualError(t, err, "migration name might contain only letters, numbers and \"_\" symbol")
	})

	t.Run("should return error if migrate command & migrations schema contains forbidden symbols", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "migrate"
		flags.MigrationsDir = "/migrations"
		flags.MigrationsTableSchema = "Hello_i contain_forbidden_symbol"
		err := flags.Validate()

		assert.EqualError(t, err, "migrations table schema might contain only letters, numbers and \"_\" symbol")
	})

	t.Run("should return error if migrate command & migrations table contains forbidden symbols", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(FS)
		flags.Command = "migrate"
		flags.MigrationsTableSchema = "jobs"
		flags.MigrationsDir = "/migrations"
		flags.MigrationsTable = "Hello_i contain_forbidden_symbol"
		err := flags.Validate()

		assert.EqualError(t, err, "migrations table might contain only letters, numbers and \"_\" symbol")
	})

	t.Run("should return error if connection string is not specified", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "migrate"
		flags.MigrationsDir = "/migrations"
		flags.MigrationsTableSchema = "jobs"
		flags.MigrationsTable = "migrations"
		err := flags.Validate()

		assert.EqualError(t, err, "connection string is required")
	})

	t.Run("should pass validation for create command", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "create"
		flags.MigrationsDir = "/migrations"
		flags.MigrationName = "initial"
		err := flags.Validate()

		assert.Nil(t, err)
	})

	t.Run("should pass validation for migrate command", func(t *testing.T) {
		flags := new(Flags)
		flags.Priority = string(DB)
		flags.Command = "migrate"
		flags.MigrationsDir = "/migrations"
		flags.MigrationsTableSchema = "jobs"
		flags.MigrationsTable = "migrations"
		flags.ConnectionString = "some connection string"
		err := flags.Validate()

		assert.Nil(t, err)
	})
}
