package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

// ApplyMigration применяет миграцию
func ApplyMigration(
	ctx context.Context,
	tx pgx.Tx,
	migrationName string,
	migrationsTableNameWithSchema string,
	upSql string,
	downSql string,
) error {
	_, err := tx.Exec(ctx, upSql)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			`INSERT INTO %s (migration_name, created_at, down_sql) VALUES ($1, CURRENT_TIMESTAMP(3), $2);`,
			migrationsTableNameWithSchema,
		),
		migrationName,
		downSql,
	); err != nil {
		return err
	}

	return nil
}
