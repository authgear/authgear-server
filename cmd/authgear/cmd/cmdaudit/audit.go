package cmdaudit

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	dbutil "github.com/authgear/authgear-server/pkg/lib/db/util"
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

var AuditMigrationSet = sqlmigrate.NewMigrateSet("_audit_migration", "migrations/audit")

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

		_, err = AuditMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, 0)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateDown = &cobra.Command{
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
			dbURL,
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
			dbURL,
			dbSchema,
			inputDir,
			args,
			tableNames,
		)

		return restorer.Restore(cmd.Context())
	},
}
