package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/quadgod/pgm/pkg/pgm/db"
	"github.com/quadgod/pgm/pkg/pgm/fs"
)

func applyWithDBPriority(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	fsMigrations []pgm.Migration,
	dbMigrations []pgm.Migration,
) ([]pgm.MigrationResult, error) {
	appliedMigrations := make([]pgm.MigrationResult, 0)

	for i, fsMigration := range fsMigrations {
		isSameDbMigrationExist := len(dbMigrations) > 0 && len(dbMigrations) > i

		// Если миграция в базе существует и она аналогична миграции
		// из файловой системы, то переходим к следующей миграции.
		if isSameDbMigrationExist && dbMigrations[i].Name == fsMigration.Name {
			continue
		}

		// Если миграция в базе не идентична миграции
		// в файловой системе, то возвращаем ошибку.
		if isSameDbMigrationExist && dbMigrations[i].Name != fsMigration.Name {
			return nil, fmt.Errorf(
				"migrations are not the same in file system and db. %s != %s",
				dbMigrations[i].Name,
				fsMigration.Name,
			)
		}

		if !isSameDbMigrationExist {
			upSqlBytes, err := os.ReadFile(fsMigration.Up)
			if err != nil {
				return nil, err
			}

			downSqlBytes, err := os.ReadFile(fsMigration.Down)
			if err != nil {
				return nil, err
			}

			err = db.ApplyMigration(
				ctx,
				tx,
				fsMigration.Name,
				migTbl,
				string(upSqlBytes),
				string(downSqlBytes),
			)

			if err != nil {
				return nil, err
			}

			appliedMigrations = append(appliedMigrations, pgm.MigrationResult{
				MigrationName: fsMigration.Name,
				Status:        pgm.APPLIED,
			})
		}
	}

	return appliedMigrations, nil
}

func applyWithFSPriority(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	fsMigrations []pgm.Migration,
	dbMigrations []pgm.Migration,
) ([]pgm.MigrationResult, error) {
	results := make([]pgm.MigrationResult, 0)

	if len(fsMigrations) == 0 {
		revertedNames, err := db.Reset(ctx, tx, migTbl)
		if err != nil {
			return nil, err
		}

		for _, name := range revertedNames {
			results = append(results, pgm.MigrationResult{
				MigrationName: name,
				Status:        pgm.REVERTED,
			})
		}

		return results, nil
	}

	for i, dbMigration := range dbMigrations {
		isFSMigrationSameIndexExist := len(fsMigrations) > 0 && len(fsMigrations) > i

		// Если миграция в базе существует и она аналогична миграции
		// из файловой системы, то переходим к следующей миграции.
		if isFSMigrationSameIndexExist && fsMigrations[i].Name == dbMigration.Name {
			continue
		}

		// Если миграция в базе не идентична миграции
		// в файловой системе, то откатываем до состояния идентичности.
		if (isFSMigrationSameIndexExist && fsMigrations[i].Name != dbMigration.Name) || !isFSMigrationSameIndexExist {
			migNamesToRevert := make([]string, 0)
			for _, dbMig := range dbMigrations[i:] {
				migNamesToRevert = append(migNamesToRevert, dbMig.Name)
			}
			revertResults, err := db.RevertMigrations(ctx, tx, migTbl, migNamesToRevert)
			if err != nil {
				return nil, err
			}
			results = append(results, revertResults...)
			break
		}
	}

	dbMigrationsAfterRevert, err := db.GetMigrations(ctx, tx, migTbl)
	if err != nil {
		return nil, err
	}

	appliedMigrations, err := applyWithDBPriority(ctx, tx, migTbl, fsMigrations, dbMigrationsAfterRevert)
	if err != nil {
		return nil, err
	}

	results = append(results, appliedMigrations...)

	return results, nil
}

func Migrate(ctx context.Context, opts *pgm.MigratorOptions) ([]pgm.MigrationResult, error) {
	fsMigrations, err := fs.ReadMigrationsDir(opts.MigrationsDir)
	if err != nil {
		return nil, err
	}

	pool, err := db.Connect(ctx, opts.ConnectionString)
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	err = db.EnsureMigrationsTable(ctx, pool, opts.MigrationsTableSchema, opts.MigrationsTable)
	if err != nil {
		return nil, err
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, rollbackErr)
		}
	}()

	err = db.LockMigrationsTable(ctx, tx, opts.MigrationsTableNameWithSchema())
	if err != nil {
		return nil, fmt.Errorf("lock migrations table error: %w", err)
	}

	dbMigrations, err := db.GetMigrations(ctx, tx, opts.MigrationsTableNameWithSchema())
	if err != nil {
		return nil, err
	}

	switch opts.Priority {
	case pgm.DB:
		result, err := applyWithDBPriority(
			ctx,
			tx,
			opts.MigrationsTableNameWithSchema(),
			fsMigrations,
			dbMigrations,
		)
		if err != nil {
			return nil, err
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}

		return result, nil
	case pgm.FS:
		result, err := applyWithFSPriority(
			ctx,
			tx,
			opts.MigrationsTableNameWithSchema(),
			fsMigrations,
			dbMigrations,
		)
		if err != nil {
			return nil, err
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}

		return result, nil
	default:
		err = fmt.Errorf("unknown priority %s", opts.Priority)
		return nil, err
	}
}
