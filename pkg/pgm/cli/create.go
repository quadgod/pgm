package cli

import (
	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/quadgod/pgm/pkg/pgm/fs"
)

func CreateMigrationFile(opts *pgm.MigratorOptions) (*pgm.Migration, error) {
	migrations, err := fs.CreateMigration(opts.MigrationsDir, opts.MigrationName)

	if err != nil {
		return nil, err
	}

	return migrations, nil
}
