package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/authgear/authgear-server/cmd/portal/migrate"
)

func init() {
	cmdDatabase.AddCommand(cmdMigrate)

	cmdMigrate.AddCommand(cmdMigrateNew)
	cmdMigrate.AddCommand(cmdMigrateUp)
	cmdMigrate.AddCommand(cmdMigrateDown)
	cmdMigrate.AddCommand(cmdMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdMigrateUp, cmdMigrateDown, cmdMigrateStatus} {
		ArgDatabaseURL.Bind(cmd.Flags(), viper.GetViper())
		ArgDatabaseSchema.Bind(cmd.Flags(), viper.GetViper())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
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
