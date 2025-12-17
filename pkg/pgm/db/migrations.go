package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/quadgod/pgm/pkg/pgm"
)

// GetMigrations считывает список миграций из базы данных
func GetMigrations(ctx context.Context, tx pgx.Tx, migTbl string) ([]pgm.Migration, error) {
	rows, err := tx.Query(ctx, fmt.Sprintf(
		`SELECT migration_name, down_sql FROM %s ORDER BY migration_name ASC;`,
		migTbl,
	))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dbMigrations := make([]pgm.Migration, 0)

	for rows.Next() {
		dbMigration := pgm.Migration{Up: ""}
		err = rows.Scan(&dbMigration.Name, &dbMigration.Down)
		if err != nil {
			return nil, err
		}
		dbMigrations = append(dbMigrations, dbMigration)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return dbMigrations, nil
}
