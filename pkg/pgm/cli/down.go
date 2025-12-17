package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/quadgod/pgm/pkg/pgm/db"
)

func Down(ctx context.Context, opts *pgm.MigratorOptions) (*pgm.MigrationResult, error) {
	pool, err := db.Connect(ctx, opts.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("database connection errors: %w", err)
	}
	defer pool.Close()

	err = db.EnsureMigrationsTable(ctx, pool, opts.MigrationsTableSchema, opts.MigrationsTable)
	if err != nil {
		return nil, fmt.Errorf("ensure migrations table errors: %w", err)
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

	if err := db.LockMigrationsTable(ctx, tx, opts.MigrationsTableNameWithSchema()); err != nil {
		return nil, err
	}

	migrations, err := db.GetMigrations(ctx, tx, opts.MigrationsTableNameWithSchema())
	if err != nil {
		return nil, err
	}

	if len(migrations) > 0 {
		_, err := tx.Exec(ctx, migrations[len(migrations)-1].Down)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, fmt.Sprintf(
			"DELETE FROM %s WHERE migration_name = $1",
			opts.MigrationsTableNameWithSchema()),
			migrations[len(migrations)-1].Name,
		)
		if err != nil {
			return nil, err
		}

		err = tx.Commit(ctx)
		if err != nil {
			return nil, fmt.Errorf("commit transaction errors: %w", err)
		}

		result := pgm.MigrationResult{
			MigrationName: migrations[len(migrations)-1].Name,
			Status:        pgm.REVERTED,
		}
		return &result, nil
	} else {
		err = tx.Commit(ctx)
		if err != nil {
			return nil, fmt.Errorf("commit transaction errors: %w", err)
		}
	}

	err = errors.New("migrations not found")

	return nil, err
}
