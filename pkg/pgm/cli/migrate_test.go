package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/quadgod/pgm/pkg/pgm/db"
	"github.com/quadgod/pgm/pkg/pgm/fs"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func genUpSql(table string) string {
	return fmt.Sprintf(`
		CREATE SCHEMA IF NOT EXISTS test;
		CREATE TABLE test.%s (
			"id"            SERIAL PRIMARY KEY,
			"active"        BOOLEAN
		);
	`, table)
}

func genDownSql(table string) string {
	return fmt.Sprintf("DROP TABLE test.%s;", table)
}

func genMigration(migrationsDir string, migrationName string, tableName string) (*pgm.Migration, error) {
	migration, err := fs.CreateMigration(migrationsDir, migrationName)

	if err != nil {
		return nil, err
	}

	upSql := genUpSql(tableName)
	if err = os.WriteFile(migration.Up, []byte(upSql), 0755); err != nil {
		return nil, err
	}

	downSql := genDownSql(tableName)
	if err = os.WriteFile(migration.Down, []byte(downSql), 0755); err != nil {
		return nil, err
	}

	return migration, nil
}

func Test_Migrate(t *testing.T) {
	ctx := context.Background()
	migrationsDir := path.Join(t.TempDir(), "/pgmigrator_migrate_test")

	postgresContainer, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
		if err := os.RemoveAll(migrationsDir); err != nil {
			t.Errorf("failed to remove migrations dir: %s", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		t.Fatalf("unable to build connection string")
	}

	t.Run("should apply 2 migrations", func(t *testing.T) {
		opts := new(pgm.MigratorOptions)
		opts.ConnectionString = connStr
		opts.Command = pgm.MIGRATE
		opts.Priority = pgm.DB
		opts.MigrationsTableSchema = "detmir_jobs"
		opts.MigrationsTable = "migrations"
		opts.MigrationsDir = migrationsDir

		_, err = genMigration(opts.MigrationsDir, "first_migration", "table1")
		if err != nil {
			t.Fatal(err)
		}

		_, err = genMigration(opts.MigrationsDir, "second_migration", "table2")
		if err != nil {
			t.Fatal(err)
		}

		appliedMigrations1, err := Migrate(ctx, opts)
		assert.Nil(t, err)
		assert.Len(t, appliedMigrations1, 2)
		assert.Contains(t, appliedMigrations1[0].MigrationName, "first_migration")
		assert.Equal(t, pgm.APPLIED, appliedMigrations1[0].Status)
		assert.Contains(t, appliedMigrations1[1].MigrationName, "second_migration")
		assert.Equal(t, pgm.APPLIED, appliedMigrations1[1].Status)

		appliedMigrations2, err := Migrate(ctx, opts)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		assert.Len(t, appliedMigrations2, 0)

		pool, err := db.Connect(ctx, connStr)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		defer pool.Close()

		migrationSchemaTables, err := db.Tables(ctx, pool, opts.MigrationsTableSchema)
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		assert.Len(t, migrationSchemaTables, 1)
		assert.Equal(t, opts.MigrationsTableSchema, migrationSchemaTables[0].TableSchema)
		assert.Equal(t, opts.MigrationsTable, migrationSchemaTables[0].TableName)

		testSchemaTables, err := db.Tables(ctx, pool, "test")
		if assert.Nil(t, err) {
			assert.Len(t, testSchemaTables, 2)
			assert.Equal(t, "table1", testSchemaTables[0].TableName)
			assert.Equal(t, "table2", testSchemaTables[1].TableName)
		}

		err = db.ResetForTests(ctx, pool, opts.MigrationsTableNameWithSchema())
		if assert.Nil(t, err) {
			testSchemaTablesAfterReset, err := db.Tables(ctx, pool, "test")
			if assert.Nil(t, err) {
				assert.Len(t, testSchemaTablesAfterReset, 0)
			}
		}

		err = os.RemoveAll(migrationsDir)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
	})

	t.Run("should test fs priority", func(t *testing.T) {
		opts := new(pgm.MigratorOptions)
		opts.ConnectionString = connStr
		opts.Command = pgm.MIGRATE
		opts.Priority = pgm.FS
		opts.MigrationsTableSchema = "detmir_jobs"
		opts.MigrationsTable = "migrations"
		opts.MigrationsDir = migrationsDir

		_, err = genMigration(opts.MigrationsDir, "first_migration", "table1")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		migration2, err := genMigration(opts.MigrationsDir, "second_migration", "table2")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		migration3, err := genMigration(opts.MigrationsDir, "third_migration", "table3")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		appliedMigrations1, err := Migrate(ctx, opts)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		assert.Len(t, appliedMigrations1, 3)

		assert.Contains(t, appliedMigrations1[0].MigrationName, "first_migration")
		assert.Equal(t, pgm.APPLIED, appliedMigrations1[0].Status)

		assert.Contains(t, appliedMigrations1[1].MigrationName, "second_migration")
		assert.Equal(t, pgm.APPLIED, appliedMigrations1[1].Status)

		assert.Contains(t, appliedMigrations1[2].MigrationName, "third_migration")
		assert.Equal(t, pgm.APPLIED, appliedMigrations1[2].Status)

		// Удалим 2 и 3 миграцию
		removeErr := errors.Join(
			os.Remove(migration2.Up),
			os.Remove(migration2.Down),
			os.Remove(migration3.Up),
			os.Remove(migration3.Down),
		)

		if !assert.Nil(t, removeErr) {
			t.FailNow()
		}

		// Создадим 4ю миграцю
		_, err = genMigration(opts.MigrationsDir, "fourth_migration", "table4")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		secondResults, err := Migrate(ctx, opts)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		assert.Len(t, secondResults, 3)
		assert.Contains(t, secondResults[0].MigrationName, "third")
		assert.Equal(t, pgm.REVERTED, secondResults[0].Status)
		assert.Contains(t, secondResults[1].MigrationName, "second")
		assert.Equal(t, pgm.REVERTED, secondResults[1].Status)
		assert.Contains(t, secondResults[2].MigrationName, "fourth")
		assert.Equal(t, pgm.APPLIED, secondResults[2].Status)

		pool, err := db.Connect(ctx, connStr)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		defer pool.Close()

		err = db.ResetForTests(ctx, pool, opts.MigrationsTableNameWithSchema())
		if assert.Nil(t, err) {
			testSchemaTablesAfterReset, err := db.Tables(ctx, pool, "test")
			if assert.Nil(t, err) {
				assert.Len(t, testSchemaTablesAfterReset, 0)
			}
		}

		err = os.RemoveAll(migrationsDir)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
	})
}
