package cmdsearch

import (
	"embed"
	"os"
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
		binder.BindString(cmd.Flags(), authgearcmd.ArgSearchDatabaseURL)
		binder.BindString(cmd.Flags(), authgearcmd.ArgSearchDatabaseSchema)
	}

	cmdSearch.AddCommand(cmdSearchReindex)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgSearchDatabaseURL)
	binder.BindString(cmdSearchReindex.Flags(), authgearcmd.ArgSearchDatabaseSchema)

	authgearcmd.Root.AddCommand(cmdSearch)
}

//go:embed migrations/search
var searchMigrationFS embed.FS

var SearchMigrationSet = sqlmigrate.NewMigrateSet(sqlmigrate.NewMigrationSetOptions{
	TableName:                            "_search_migration",
	EmbedFS:                              searchMigrationFS,
	EmbedFSRoot:                          "migrations/search",
	OutputPathRelativeToWorkingDirectory: "./cmd/authgear/cmd/cmdsearch/migrations/search",
})

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
	Use:   sqlmigrate.CobraMigrateUpUse,
	Short: sqlmigrate.CobraMigrateUpShort,
	Args:  sqlmigrate.CobraMigrateUpArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseSchema)
		if err != nil {
			return
		}

		n, err := sqlmigrate.CobraParseMigrateUpArgs(args)
		if err != nil {
			return
		}

		_, err = SearchMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, n)
		if err != nil {
			return
		}

		return
	},
}

var cmdSearchDatabaseMigrateDown = &cobra.Command{
	Hidden: true,
	Use:    sqlmigrate.CobraMigrateDownUse,
	Short:  sqlmigrate.CobraMigrateDownShort,
	Args:   sqlmigrate.CobraMigrateDownArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseSchema)
		if err != nil {
			return
		}

		numMigrations, err := sqlmigrate.CobraParseMigrateDownArgs(args)
		if err != nil {
			return
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
	Use:   sqlmigrate.CobraMigrateStatusUse,
	Short: sqlmigrate.CobraMigrateStatusShort,
	Args:  sqlmigrate.CobraMigrateStatusArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgSearchDatabaseSchema)
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
