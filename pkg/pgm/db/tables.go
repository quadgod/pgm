package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TableInfo struct {
	TableSchema string
	TableName   string
}

// Tables получает список таблиц схемы
func Tables(ctx context.Context, pool *pgxpool.Pool, schema string) ([]TableInfo, error) {
	rows, err := pool.Query(
		ctx,
		`SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema = $1`,
		schema,
	)

	if err != nil {
		return nil, err
	}

	tables, err := pgx.CollectRows(rows, pgx.RowToStructByName[TableInfo])
	if err != nil {
		return nil, fmt.Errorf("collect table names rows error: %v", err)
	}

	return tables, nil
}
