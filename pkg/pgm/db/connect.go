package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
)

func getDatabaseNameFromConnectionString(connectionString string) (string, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	return config.ConnConfig.Database, nil
}

func checkDatabaseExists(ctx context.Context, conn *pgxpool.Pool, database string) (bool, error) {
	var exists bool
	err := conn.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		database,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func createDatabase(ctx context.Context, conn *pgxpool.Pool, database string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

func preconnect(ctx context.Context, connectionString string) error {
	database, err := getDatabaseNameFromConnectionString(connectionString)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	if database == "" {
		return nil
	}

	connStringWithoutDb := removeDatabaseFromConnectionString(connectionString, database)
	dbCfg, err := pgxpool.ParseConfig(connStringWithoutDb)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}
	dbCfg.ConnConfig.StatementCacheCapacity = 0
	dbCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("database \"%s\" connection error: %w", database, err)
	}
	defer conn.Close()

	exists, err := checkDatabaseExists(ctx, conn, database)
	if err != nil {
		return fmt.Errorf("failed to check database \"%s\" existence: %w", database, err)
	}

	if !exists {
		err = createDatabase(ctx, conn, database)
		if err != nil {
			return fmt.Errorf("database \"%s\" creation error: %w", database, err)
		}
	}

	return nil
}

func removeDatabaseFromConnectionString(connectionString, database string) string {
	re := regexp.MustCompile(`(.*)/(` + regexp.QuoteMeta(database) + `)(\?.*)?$`)
	return re.ReplaceAllString(connectionString, "$1$3")
}

func Connect(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	err := preconnect(ctx, connectionString)
	if err != nil {
		fmt.Printf("Preconnect warning: %v (attempting to connect anyway)\n", err)
	}

	conn, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return conn, nil
}
