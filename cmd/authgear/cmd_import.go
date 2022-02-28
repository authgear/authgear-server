package main

import (
	"context"

	"github.com/spf13/cobra"

	cmdimporter "github.com/authgear/authgear-server/cmd/authgear/importer"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var cmdImport = &cobra.Command{
	Use:   "import [csv path]",
	Short: "Import external users",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := getBinder()

		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		appID, err := binder.GetRequiredString(cmd, ArgAppID)
		if err != nil {
			return err
		}

		emailCaseSensitive, _ := cmd.Flags().GetBool("email-case-sensitive")
		emailBlockPlusSign, _ := cmd.Flags().GetBool("email-block-plus-sign")
		emailIgnoreDotSign, _ := cmd.Flags().GetBool("email-ignore-dot-sign")

		loginIDEmailConfig := &config.LoginIDEmailConfig{
			CaseSensitive: &emailCaseSensitive,
			BlockPlusSign: &emailBlockPlusSign,
			IgnoreDotSign: &emailIgnoreDotSign,
		}
		loginIDEmailConfig.SetDefaults()

		csvPath := args[0]

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		dbPool := db.NewPool()

		importer := cmdimporter.NewImporter(
			context.Background(),
			dbPool,
			dbCredentials,
			config.AppID(appID),
			loginIDEmailConfig,
		)

		return importer.ImportFromCSV(csvPath)
	},
}

func init() {
	binder := getBinder()
	binder.BindString(
		cmdImport.PersistentFlags(),
		ArgDatabaseSchema,
	)
	binder.BindString(
		cmdImport.PersistentFlags(),
		ArgDatabaseURL,
	)
	binder.BindString(
		cmdImport.PersistentFlags(),
		ArgAppID,
	)
	_ = cmdImport.Flags().Bool("email-case-sensitive", false, "set email is case sensitive")
	_ = cmdImport.Flags().Bool("email-block-plus-sign", false, "disallow plus sign in email")
	_ = cmdImport.Flags().Bool("email-ignore-dot-sign", false, "ignore the dot sign in email")
}
