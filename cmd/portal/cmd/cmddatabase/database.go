package cmddatabase

import (
	"embed"
	"os"
	"strings"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	dbutil "github.com/authgear/authgear-server/pkg/lib/db/util"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	binder := portalcmd.GetBinder()
	cmdDatabase.AddCommand(cmdMigrate)
	cmdDatabase.AddCommand(cmdDump)
	cmdDatabase.AddCommand(cmdRestore)

	cmdMigrate.AddCommand(cmdMigrateNew)
	cmdMigrate.AddCommand(cmdMigrateUp)
	cmdMigrate.AddCommand(cmdMigrateDown)
	cmdMigrate.AddCommand(cmdMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdMigrateUp, cmdMigrateDown, cmdMigrateStatus} {
		binder.BindString(cmd.Flags(), portalcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), portalcmd.ArgDatabaseSchema)
	}

	binder.BindString(cmdDump.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdDump.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdDump.Flags(), portalcmd.ArgOutputDirectoryPath)

	binder.BindString(cmdRestore.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdRestore.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdRestore.Flags(), portalcmd.ArgInputDirectoryPath)

	portalcmd.Root.AddCommand(cmdDatabase)
}

//go:embed migrations/portal
var portalMigrationFS embed.FS

var PortalMigrationSet = sqlmigrate.NewMigrateSet(sqlmigrate.NewMigrationSetOptions{
	TableName:                            "_portal_migration",
	EmbedFS:                              portalMigrationFS,
	EmbedFSRoot:                          "migrations/portal",
	OutputPathRelativeToWorkingDirectory: "./cmd/portal/cmd/cmddatabase/migrations/portal",
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
		_, err = PortalMigrationSet.Create(name)
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		n, err := sqlmigrate.CobraParseMigrateUpArgs(args)
		if err != nil {
			return
		}

		_, err = PortalMigrationSet.Up(sqlmigrate.ConnectionOptions{
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		numMigrations, err := sqlmigrate.CobraParseMigrateDownArgs(args)
		if err != nil {
			return
		}

		_, err = PortalMigrationSet.Down(sqlmigrate.ConnectionOptions{
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}

		plans, err := PortalMigrationSet.Status(sqlmigrate.ConnectionOptions{
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}
		outputDir, err := binder.GetRequiredString(cmd, portalcmd.ArgOutputDirectoryPath)
		if err != nil {
			return
		}

		if len(args) == 0 {
			os.Exit(0)
		}

		dumper := dbutil.NewDumper(
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeGlobal,
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return
		}
		inputDir, err := binder.GetRequiredString(cmd, portalcmd.ArgInputDirectoryPath)
		if err != nil {
			return
		}

		restorer := dbutil.NewRestorer(
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeGlobal,
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
