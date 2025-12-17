package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Reset удаляет все миграции из базы данных
func Reset(
	ctx context.Context,
	tx pgx.Tx,
	migrationsTableNameWithSchema string,
) ([]string, error) {
	revertedNames := make([]string, 0)

	dbMigrations, err := GetMigrations(ctx, tx, migrationsTableNameWithSchema)
	if err != nil {
		return nil, err
	}

	if len(dbMigrations) < 1 {
		return revertedNames, nil
	}

	for i := len(dbMigrations) - 1; i >= 0; i-- {
		_, err = tx.Exec(ctx, dbMigrations[i].Down)
		if err != nil {
			return nil, err
		}
		revertedNames = append(revertedNames, dbMigrations[i].Name)
	}

	_, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s;", migrationsTableNameWithSchema))
	if err != nil {
		return nil, err
	}

	return revertedNames, err
}

// ResetForTests откатывает все миграции (используется только в тестах)
func ResetForTests(
	ctx context.Context,
	pool *pgxpool.Pool,
	migrationsTableNameWithSchema string,
) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, rollbackErr)
		}
	}()

	dbMigrations, err := GetMigrations(ctx, tx, migrationsTableNameWithSchema)
	if err != nil {
		return err
	}

	if len(dbMigrations) < 1 {
		return nil
	}

	for i := len(dbMigrations) - 1; i >= 0; i-- {
		_, err = tx.Exec(ctx, dbMigrations[i].Down)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s;", migrationsTableNameWithSchema))
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return err
}
