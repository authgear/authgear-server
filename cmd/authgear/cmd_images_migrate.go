package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	binder := getBinder()
	cmdImages.AddCommand(cmdImagesDatabase)
	cmdImagesDatabase.AddCommand(cmdImagesDatabaseMigrate)

	cmdImagesDatabaseMigrate.AddCommand(cmdImagesDatabaseMigrateNew)
	cmdImagesDatabaseMigrate.AddCommand(cmdImagesDatabaseMigrateUp)
	cmdImagesDatabaseMigrate.AddCommand(cmdImagesDatabaseMigrateDown)
	cmdImagesDatabaseMigrate.AddCommand(cmdImagesDatabaseMigrateStatus)

	for _, cmd := range []*cobra.Command{
		cmdImagesDatabaseMigrateUp,
		cmdImagesDatabaseMigrateDown,
		cmdImagesDatabaseMigrateStatus,
	} {
		binder.BindString(cmd.Flags(), ArgDatabaseURL)
		binder.BindString(cmd.Flags(), ArgDatabaseSchema)
	}
}

var ImagesMigrationSet = sqlmigrate.NewMigrateSet("_images_migrations", "migrations/images")

var cmdImagesDatabase = &cobra.Command{
	Use:   "database migrate",
	Short: "Images database commands",
}

var cmdImagesDatabaseMigrate = &cobra.Command{
	Use:   "migrate [new|status|up|down]",
	Short: "Migrate database schema",
}

var cmdImagesDatabaseMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := strings.Join(args, "_")
		_, err = ImagesMigrationSet.Create(name)
		if err != nil {
			return
		}
		return
	},
}

var cmdImagesDatabaseMigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Migrate database schema to latest version",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return
		}
		_, err = ImagesMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, 0)
		if err != nil {
			return
		}
		return
	},
}

var cmdImagesDatabaseMigrateDown = &cobra.Command{
	Use:    "down",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return
		}

		if len(args) == 0 {
			err = fmt.Errorf("number of migrations to revert not specified; specify 'all' to revert all migrations")
			return
		}

		var numMigrations int
		if args[0] == "all" {
			numMigrations = 0
		} else {
			numMigrations, err = strconv.Atoi(args[0])
			if err != nil {
				err = fmt.Errorf("invalid number of migrations specified: %s", err)
				return
			} else if numMigrations <= 0 {
				err = fmt.Errorf("no migrations specified to revert")
				return
			}
		}

		_, err = ImagesMigrationSet.Down(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, numMigrations)
		if err != nil {
			return
		}

		return
	},
}

var cmdImagesDatabaseMigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Get database schema migration status",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return
		}
		plans, err := ImagesMigrationSet.Status(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		})
		if err != nil {
			return
		}
		if len(plans) != 0 {
			os.Exit(1)
		}
		return
	},
}
