package cmddatabase

import (
	"embed"
	"os"
	"strings"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimages"
	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	binder := authgearcmd.GetBinder()
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
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseURL)
		binder.BindString(cmd.Flags(), authgearcmd.ArgDatabaseSchema)
	}

	cmdimages.CmdImages.AddCommand(cmdImagesDatabase)
}

//go:embed migrations/images
var imagesMigrationFS embed.FS

var ImagesMigrationSet = sqlmigrate.NewMigrateSet(sqlmigrate.NewMigrationSetOptions{
	TableName:                            "_images_migrations",
	EmbedFS:                              imagesMigrationFS,
	EmbedFSRoot:                          "migrations/images",
	OutputPathRelativeToWorkingDirectory: "./cmd/authgear/cmd/cmdimages/cmddatabase/migrations/images",
})

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
		_, err = ImagesMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, n)
		if err != nil {
			return
		}
		return
	},
}

var cmdImagesDatabaseMigrateDown = &cobra.Command{
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
