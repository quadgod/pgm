package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureMigrationsTable создает схему и таблицу миграций если они не существуют
func EnsureMigrationsTable(
	ctx context.Context,
	conn *pgxpool.Pool,
	migrationsTableSchemaName string,
	migrationsTableName string,
) error {
	_, err := conn.Exec(ctx,
		fmt.Sprintf(`
			CREATE SCHEMA IF NOT EXISTS %s;
			CREATE TABLE IF NOT EXISTS %s.%s (
				migration_name VARCHAR(512) NOT NULL,
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP(3),
				down_sql TEXT NOT NULL,
				CONSTRAINT "%s_pk" PRIMARY KEY (migration_name)
			);
		`,
			migrationsTableSchemaName,
			migrationsTableSchemaName,
			migrationsTableName,
			migrationsTableName,
		),
	)

	if err != nil && err.Error() != pgx.ErrNoRows.Error() {
		return err
	}

	return nil
}
