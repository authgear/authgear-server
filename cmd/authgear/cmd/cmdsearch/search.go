package cmdsearch

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	binder := authgearcmd.GetBinder()
	cmdSearch.AddCommand(cmdSearchDatabase)
	cmdSearchDatabase.AddCommand(cmdSearchDatabaseMigrate)

	cmdSearchDatabaseMigrate.AddCommand(cmdSearchDatabaseMigrateNew)
	cmdSearchDatabaseMigrate.AddCommand(cmdSearchDatabaseMigrateUp)
	cmdSearchDatabaseMigrate.AddCommand(cmdSearchDatabaseMigrateDown)
	cmdSearchDatabaseMigrate.AddCommand(cmdSearchDatabaseMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdSearchDatabaseMigrateUp, cmdSearchDatabaseMigrateDown, cmdSearchDatabaseMigrateStatus} {
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseSchema)
	}

	cmdSearch.AddCommand(cmdSearchReindex)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgSearchDatabaseURL)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgSearchDatabaseSchema)

	authgearcmd.Root.AddCommand(cmdSearch)
}

var SearchMigrationSet = sqlmigrate.NewMigrateSet("_search_migration", "migrations/search")

var cmdSearch = &cobra.Command{
	Use:   "search",
	Short: "Search commands",
}

var cmdSearchDatabase = &cobra.Command{
	Use:   "database [migrate]",
	Short: "Search database commands",
}

var cmdSearchDatabaseMigrate = &cobra.Command{
	Use:   "migrate [new|status|up|down]",
	Short: "Migrate database schema",
}

var cmdSearchDatabaseMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := strings.Join(args, "_")
		_, err = SearchMigrationSet.Create(name)
		if err != nil {
			return
		}

		return
	},
}

var cmdSearchDatabaseMigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Migrate database schema to latest version",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		_, err = SearchMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, 0)
		if err != nil {
			return
		}

		return
	},
}

var cmdSearchDatabaseMigrateDown = &cobra.Command{
	Use:    "down",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
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

		_, err = SearchMigrationSet.Down(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, numMigrations)
		if err != nil {
			return
		}

		return
	},
}

var cmdSearchDatabaseMigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Get database schema migration status",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		plans, err := SearchMigrationSet.Status(sqlmigrate.ConnectionOptions{
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
