package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/quadgod/pgm/pkg/pgm"
)

// RevertMigration откатывает миграцию
func RevertMigration(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	migName string,
) (*pgm.MigrationResult, error) {
	downSqlQuery := fmt.Sprintf(`SELECT down_sql FROM %s WHERE migration_name = $1 LIMIT 1;`, migTbl)
	row := tx.QueryRow(ctx, downSqlQuery, migName)

	var downSql string
	if err := row.Scan(&downSql); err != nil {
		return nil, fmt.Errorf("revert %s migration error - down sql was not found. %v", migName, err)
	}

	if _, err := tx.Exec(ctx, downSql); err != nil {
		return nil, fmt.Errorf("revert %s migration error - can't execute down sql. %v", migName, err)
	}

	delMigSql := fmt.Sprintf(`DELETE FROM %s WHERE migration_name = $1;`, migTbl)
	if _, err := tx.Exec(ctx, delMigSql, migName); err != nil {
		return nil, fmt.Errorf("revert %s migration error - can't delete record from migrations table. %v", migName, err)
	}

	result := new(pgm.MigrationResult)
	result.MigrationName = migName
	result.Status = pgm.REVERTED

	return result, nil
}

// RevertMigrations откатывает набор миграций в обратном порядке
func RevertMigrations(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	migNames []string,
) ([]pgm.MigrationResult, error) {
	results := make([]pgm.MigrationResult, 0)

	for i := len(migNames) - 1; i >= 0; i-- {
		result, err := RevertMigration(ctx, tx, migTbl, migNames[i])
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	return results, nil
}
