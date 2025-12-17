package fs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func Test_CreateMigration(t *testing.T) {
	migrationsDir := path.Join(t.TempDir(), "/filesystem_create_test")

	t.Cleanup(func() {
		err := os.RemoveAll(migrationsDir)
		if err != nil {
			t.Fatalf("unable to cleanup migrations dir %s", err)
		}
	})

	t.Run("should create migrations files", func(t *testing.T) {
		migrationDirEntries1, err := os.ReadDir(migrationsDir)
		assert.Nil(t, migrationDirEntries1)
		assert.True(t, os.IsNotExist(err))

		migrations1, err := CreateMigration(migrationsDir, "first_migration")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		assert.Contains(t, migrations1.Up, ".up.sql")
		assert.Contains(t, migrations1.Down, ".down.sql")

		migrations2, err := CreateMigration(migrationsDir, "second_migration")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		assert.Contains(t, migrations2.Up, ".up.sql")
		assert.Contains(t, migrations2.Down, ".down.sql")

		migrationDirEntries2, err := os.ReadDir(migrationsDir)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		assert.Len(t, migrationDirEntries2, 4)
	})
}
