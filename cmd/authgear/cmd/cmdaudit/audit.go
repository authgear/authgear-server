package cmdaudit

import (
	"embed"
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
	cmdAudit.AddCommand(cmdAuditDatabase)
	cmdAuditDatabase.AddCommand(cmdAuditDatabaseMigrate)
	cmdAuditDatabase.AddCommand(cmdAuditDatabaseMaintain)
	cmdAuditDatabase.AddCommand(cmdAuditDatabaseDump)
	cmdAuditDatabase.AddCommand(cmdAuditDatabaseRestore)

	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateNew)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateUp)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateDown)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdAuditDatabaseMigrateUp, cmdAuditDatabaseMigrateDown, cmdAuditDatabaseMigrateStatus, cmdAuditDatabaseMaintain} {
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseSchema)
	}

	binder.BindString(cmdAuditDatabaseDump.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdAuditDatabaseDump.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdAuditDatabaseDump.Flags(), authgearcmd.ArgOutputFolder)

	binder.BindString(cmdAuditDatabaseRestore.Flags(), authgearcmd.ArgDatabaseURL)
	binder.BindString(cmdAuditDatabaseRestore.Flags(), authgearcmd.ArgDatabaseSchema)
	binder.BindString(cmdAuditDatabaseRestore.Flags(), authgearcmd.ArgInputFolder)

	authgearcmd.Root.AddCommand(cmdAudit)
}

//go:embed migrations/audit
var auditMigrationFS embed.FS

var AuditMigrationSet = sqlmigrate.NewMigrateSet(sqlmigrate.NewMigrationSetOptions{
	TableName:                            "_audit_migration",
	EmbedFS:                              auditMigrationFS,
	EmbedFSRoot:                          "migrations/audit",
	OutputPathRelativeToWorkingDirectory: "./cmd/authgear/cmd/cmdaudit/migrations/audit",
})

var cmdAudit = &cobra.Command{
	Use:   "audit database",
	Short: "Audit log commands",
}

var cmdAuditDatabase = &cobra.Command{
	Use:   "database [migrate|maintain]",
	Short: "Audit log database commands",
}

var cmdAuditDatabaseMigrate = &cobra.Command{
	Use:   "migrate [new|status|up|down]",
	Short: "Migrate database schema",
}

var cmdAuditDatabaseMaintain = &cobra.Command{
	Use:   "maintain",
	Short: "Run pg_partman maintain procedure",
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

		maintainer := sqlmigrate.PartmanMaintainer{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			TableName:      "_audit_log",
		}
		err = maintainer.RunMaintenance()
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := strings.Join(args, "_")
		_, err = AuditMigrationSet.Create(name)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateUp = &cobra.Command{
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

		_, err = AuditMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, n)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateDown = &cobra.Command{
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

		_, err = AuditMigrationSet.Down(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, numMigrations)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateStatus = &cobra.Command{
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

		plans, err := AuditMigrationSet.Status(sqlmigrate.ConnectionOptions{
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

var cmdAuditDatabaseDump = &cobra.Command{
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
			os.Exit(0)
		}

		dumper := dbutil.NewDumper(
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeAuditReadOnly,
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

var cmdAuditDatabaseRestore = &cobra.Command{
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
				Purpose:     db.ConnectionPurposeAuditReadWrite,
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
