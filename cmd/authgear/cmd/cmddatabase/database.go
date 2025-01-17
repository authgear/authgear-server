package cmddatabase

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	dbutil "github.com/authgear/authgear-server/pkg/lib/db/util"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	binder := authgearcmd.GetBinder()
	cmdDatabase.AddCommand(cmdMigrate)
	cmdDatabase.AddCommand(cmdDump)
	cmdDatabase.AddCommand(cmdRestore)

	cmdMigrate.AddCommand(cmdMigrateNew)
	cmdMigrate.AddCommand(cmdMigrateUp)
	cmdMigrate.AddCommand(cmdMigrateDown)
	cmdMigrate.AddCommand(cmdMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdMigrateUp, cmdMigrateDown, cmdMigrateStatus} {
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseSchema)
	}

	binder.BindString(cmdDump.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdDump.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdDump.Flags(), authgearcmd.ArgOutputFolder)

	binder.BindString(cmdRestore.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdRestore.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdRestore.Flags(), authgearcmd.ArgInputFolder)

	authgearcmd.Root.AddCommand(cmdDatabase)
}

//go:embed migrations/authgear
var mainMigrationFS embed.FS

var MainMigrationSet = sqlmigrate.NewMigrateSet(sqlmigrate.NewMigrationSetOptions{
	TableName:                            "_auth_migration",
	EmbedFS:                              mainMigrationFS,
	EmbedFSRoot:                          "migrations/authgear",
	OutputPathRelativeToWorkingDirectory: "./cmd/authgear/cmd/cmddatabase/migrations/authgear",
})

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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := strings.Join(args, "_")
		_, err = MainMigrationSet.Create(name)
		if err != nil {
			return
		}

		return
	},
}

var cmdMigrateUp = &cobra.Command{
	Use:   sqlmigrate.CobraMigrateUpUse,
	Short: sqlmigrate.CobraMigrateUpShort,
	Args:  sqlmigrate.CobraMigrateUpArgs,
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

		n, err := sqlmigrate.CobraParseMigrateUpArgs(args)
		if err != nil {
			return
		}

		_, err = MainMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, n)
		if err != nil {
			return
		}

		return
	},
}

var cmdMigrateDown = &cobra.Command{
	Hidden: true,
	Use:    sqlmigrate.CobraMigrateDownUse,
	Short:  sqlmigrate.CobraMigrateDownShort,
	Args:   sqlmigrate.CobraMigrateDownArgs,
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

		numMigrations, err := sqlmigrate.CobraParseMigrateDownArgs(args)
		if err != nil {
			return
		}

		_, err = MainMigrationSet.Down(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, numMigrations)
		if err != nil {
			return
		}

		return
	},
}

var cmdMigrateStatus = &cobra.Command{
	Use:   sqlmigrate.CobraMigrateStatusUse,
	Short: sqlmigrate.CobraMigrateStatusShort,
	Args:  sqlmigrate.CobraMigrateStatusArgs,
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

		plans, err := MainMigrationSet.Status(sqlmigrate.ConnectionOptions{
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

var cmdDump = &cobra.Command{
	Use:   "dump [app-id ...]",
	Short: "Dump app database into csv files.",
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
		outputDir, err := binder.GetRequiredString(cmd, authgearcmd.ArgOutputFolder)
		if err != nil {
			return
		}

		if len(args) == 0 {
			panic(fmt.Errorf("At least 1 app-id is needed."))
		}

		dumper := dbutil.NewDumper(
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeApp,
				DatabaseURL: dbURL,
			},
			dbSchema,
			outputDir,
			args,
			tableNames,
		)

		return dumper.Dump(cmd.Context())
	},
}

var cmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Restore csv files into database.",
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
		inputDir, err := binder.GetRequiredString(cmd, authgearcmd.ArgInputFolder)
		if err != nil {
			return
		}

		restorer := dbutil.NewRestorer(
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeApp,
				DatabaseURL: dbURL,
			},
			dbSchema,
			inputDir,
			args,
			tableNames,
		)

		return restorer.Restore(cmd.Context())
	},
}
