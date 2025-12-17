package fs

import (
	"os"
	"path"
	"testing"

	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/stretchr/testify/assert"
)

func Test_ReadMigrationsDir(t *testing.T) {
	migrationsDir := path.Join(t.TempDir(), "/pgmigrator_read_migrations_dir_test")

	t.Cleanup(func() {
		if err := os.RemoveAll(migrationsDir); err != nil {
			t.Error(err)
		}
	})

	t.Run("should not read migration file because it's contains forbidden symbols", func(t *testing.T) {
		forbiddenMigrationsDir := path.Join(migrationsDir, "/forbidden")
		invalidFileNamePath := path.Join(forbiddenMigrationsDir, "1_invalid-name.down.sql")

		err := os.MkdirAll(forbiddenMigrationsDir, 0755)
		if !assert.Nil(t, err) {
			assert.FailNow(t, err.Error())
		}

		f, err := os.Create(invalidFileNamePath)
		if !assert.Nil(t, err) {
			assert.FailNow(t, err.Error())
		}
		err = f.Close()

		if !assert.Nil(t, err) {
			assert.FailNow(t, err.Error())
		}

		migrations, err := ReadMigrationsDir(forbiddenMigrationsDir)
		if !assert.Nil(t, migrations) {
			t.FailNow()
		}
		assert.ErrorContains(t, err, "invalid migration file name found \"1_invalid-name.down.sql\"")

		t.Cleanup(func() {
			if err := os.RemoveAll(forbiddenMigrationsDir); err != nil {
				t.Error(err)
			}
		})
	})

	t.Run("should not read because migrations dir not exist", func(t *testing.T) {
		migrations, err := ReadMigrationsDir(path.Join(migrationsDir, "/not_exists"))
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "migrations directory does not exist")
	})

	t.Run("should create and read migrations", func(t *testing.T) {
		opts := new(pgm.MigratorOptions)
		opts.MigrationsDir = migrationsDir
		opts.MigrationName = "initial"

		err := os.MkdirAll(path.Join(migrationsDir, "/test"), 0750)
		if err != nil {
			t.Fatal(err)
		}

		txtFile, err := os.Create(path.Join(migrationsDir, "test.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer func(txtFile *os.File) {
			_ = txtFile.Close()
		}(txtFile)

		m1, err := CreateMigration(migrationsDir, "table1")
		if err != nil {
			t.Fatal(err)
		}

		m2, err := CreateMigration(migrationsDir, "table2")
		if err != nil {
			t.Fatal(err)
		}

		migrations, err := ReadMigrationsDir(opts.MigrationsDir)
		assert.Nil(t, err)

		assert.Equal(t, m1.Name, migrations[0].Name)
		assert.Equal(t, m1.Up, migrations[0].Up)
		assert.Equal(t, m1.Down, migrations[0].Down)

		assert.Equal(t, m2.Name, migrations[1].Name)
		assert.Equal(t, m2.Up, migrations[1].Up)
		assert.Equal(t, m2.Down, migrations[1].Down)
	})
}

func Test_group(t *testing.T) {
	migrationsDir := "/migrations"

	t.Run("should group empty slice of migrations files", func(t *testing.T) {
		migrationFiles := make([]string, 0)
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(migrations))
	})

	t.Run("should group migrations files", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql", "1_initial.up.sql", "2_test.down.sql", "2_test.up.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, err)
		assert.Len(t, migrations, 2)
		assert.Equal(t, "1_initial", migrations[0].Name)
		assert.Equal(t, "2_test", migrations[1].Name)
		assert.Equal(t, "/migrations/1_initial.up.sql", migrations[0].Up)
		assert.Equal(t, "/migrations/1_initial.down.sql", migrations[0].Down)
		assert.Equal(t, "/migrations/2_test.up.sql", migrations[1].Up)
		assert.Equal(t, "/migrations/2_test.down.sql", migrations[1].Down)
	})

	t.Run("should not group because no up migration for 1_initial.down.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql", "2_test.down.sql", "2_test.up.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found up migration for \"1_initial.down.sql\" migration")
	})

	t.Run("should not group because no up migration for 2_test.down.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql", "1_initial.up.sql", "2_test.down.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found up migration for \"2_test.down.sql\" migration")
	})

	t.Run("should not group because no down migration for 1_initial.up.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.up.sql", "2_test.down.sql", "2_test.up.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found down migration for \"1_initial.up.sql\" migration")
	})

	t.Run("should not group because no down migration for 2_test.up.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql", "1_initial.up.sql", "2_test.up.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found down migration for \"2_test.up.sql\" migration")
	})

	t.Run("should not group because no up migration for 1_initial.down.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found up migration for \"1_initial.down.sql\" migration")
	})

	t.Run("should not group because invalid file name", func(t *testing.T) {
		migrationFiles := []string{"1_initial.dawn.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "invalid migration filename \"1_initial.dawn.sql\"")
	})

	t.Run("should not group because no up migration for 1_initial.down.sql", func(t *testing.T) {
		migrationFiles := []string{"1_initial.down.sql", "2_test.up.sql"}
		migrations, err := group(migrationFiles, migrationsDir)
		assert.Nil(t, migrations)
		assert.ErrorContains(t, err, "not found up migration for \"1_initial.down.sql\"")
	})
}
