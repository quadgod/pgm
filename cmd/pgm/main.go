package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/quadgod/pgm/pkg/pgm"
	"github.com/quadgod/pgm/pkg/pgm/cli"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	flags := new(pgm.Flags)

	showVersion := flag.Bool("version", false, "show version and exit")
	flag.StringVar(&flags.Command, "command", "", "command")
	flag.StringVar(&flags.MigrationsDir, "migrationsDir", "", "path to migrations directory")
	flag.StringVar(&flags.MigrationsTable, "migrationsTable", "migrations", "name of migrations table")
	flag.StringVar(&flags.MigrationsTableSchema, "migrationsTableSchema", "", "migrations table schema")
	flag.StringVar(&flags.MigrationName, "migrationName", "", "migration name")
	flag.StringVar(&flags.ConnectionString, "connectionString", os.Getenv("PG_CONNECTION_STRING"), "connection string")
	flag.StringVar(&flags.Priority, "priority", string(pgm.FS), "db or fs migrations priority")

	flag.Parse()

	if *showVersion {
		fmt.Printf("pgm v%s\n", pgm.Version)
		os.Exit(0)
		return
	}

	if err := flags.Validate(); err != nil {
		logger.Error("arguments validation error", "error", err)
		os.Exit(1)
		return
	}

	opts := flags.ToMigratorOptions()

	switch opts.Command {
	case pgm.CREATE:
		mig, err := cli.CreateMigrationFile(&opts)

		if err != nil {
			logger.Error("create migration error", "error", err)
			os.Exit(1)
			return
		}

		logger.Info("migration files created", "up", mig.Up, "down", mig.Down)
	case pgm.MIGRATE:
		res, err := cli.Migrate(context.Background(), &opts)
		if err != nil {
			logger.Error("error occurs during migrate command execution", "error", err)
			os.Exit(1)
			return
		}

		for _, r := range res {
			logger.Info(r.MigrationName, "status", r.Status)
		}
	case pgm.DOWN:
		res, err := cli.Down(context.Background(), &opts)
		if err != nil {
			logger.Error("error occurs during down command execution", "error", err)
			os.Exit(1)
			return
		}

		logger.Info(res.MigrationName, "status", res.Status)
	}
}
