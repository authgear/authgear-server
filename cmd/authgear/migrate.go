package main

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/migrate"
)

var DatabaseURL string
var DatabaseSchema string

func init() {
	cmdDatabase.AddCommand(cmdMigrate)

	cmdMigrate.AddCommand(cmdMigrateNew)
	cmdMigrate.AddCommand(cmdMigrateUp)
	cmdMigrate.AddCommand(cmdMigrateDown)
	cmdMigrate.AddCommand(cmdMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdMigrateUp, cmdMigrateDown, cmdMigrateStatus} {
		cmd.Flags().StringVar(
			&DatabaseURL,
			"database-url",
			"",
			"Database URL",
		)
		cmd.Flags().StringVar(
			&DatabaseSchema,
			"database-schema",
			"",
			"Database schema name",
		)
	}
}

var cmdDatabase = &cobra.Command{
	Use:   "database migrate",
	Short: "Database commands",
}

var cmdMigrate = &cobra.Command{
	Use:   "migrate [new|status|up|down]",
	Short: "Migrate database schema",
}

var cmdMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.Join(args, "_")
		migrate.CreateMigration(name)
	},
}

var cmdMigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Migrate database schema to latest version",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		migrate.Up(migrate.Options{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		})
	},
}

var cmdMigrateDown = &cobra.Command{
	Use:    "down",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		if len(args) == 0 {
			log.Fatalf("number of migrations to revert not specified; specify 'all' to revert all migrations")
		}

		var numMigrations int
		if args[0] == "all" {
			numMigrations = 0
		} else {
			numMigrations, err = strconv.Atoi(args[0])
			if err != nil {
				log.Fatalf("invalid number of migrations specified: %s", err)
			} else if numMigrations <= 0 {
				log.Fatal("no migrations specified to revert")
			}
		}

		migrate.Down(numMigrations, migrate.Options{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		})
	},
}

var cmdMigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Get database schema migration status",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("cannot load secret config: %s", err)
		}

		latest := migrate.Status(migrate.Options{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		})
		if !latest {
			os.Exit(1)
		}
	},
}

func loadDBCredentials() (dbURL string, dbSchema string, err error) {
	if DatabaseURL == "" {
		DatabaseURL = os.Getenv("DATABASE_URL")
	}
	if DatabaseSchema == "" {
		DatabaseSchema = os.Getenv("DATABASE_SCHEMA")
	}

	if DatabaseURL == "" {
		return "", "", errors.New("missing database URL")
	}
	if DatabaseSchema == "" {
		return "", "", errors.New("missing database schema")
	}
	return DatabaseURL, DatabaseSchema, nil
}
