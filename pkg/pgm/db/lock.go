package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

// LockMigrationsTable блокирует изменения в таблице миграций в рамках транзакции
func LockMigrationsTable(ctx context.Context, tx pgx.Tx, migrationsTableNameWithSchema string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(
		"LOCK TABLE %s IN ACCESS EXCLUSIVE MODE",
		migrationsTableNameWithSchema,
	))

	return err
}
