package cmdimport

import (
	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdimporter "github.com/authgear/authgear-server/cmd/authgear/importer"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var cmdImport = &cobra.Command{
	Use:   "import [csv path]",
	Short: "Import external users",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		binder := authgearcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, authgearcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		appID, err := binder.GetRequiredString(cmd, authgearcmd.ArgAppID)
		if err != nil {
			return err
		}

		emailCaseSensitive, _ := cmd.Flags().GetBool("email-case-sensitive")
		emailBlockPlusSign, _ := cmd.Flags().GetBool("email-block-plus-sign")
		emailIgnoreDotSign, _ := cmd.Flags().GetBool("email-ignore-dot-sign")
		emailMarkAsVerified, _ := cmd.Flags().GetBool("email-mark-as-verified")

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
			dbPool,
			dbCredentials,
			config.AppID(appID),
			loginIDEmailConfig,
		)

		opts := cmdimporter.ImportOptions{
			EmailMarkAsVerified: emailMarkAsVerified,
		}

		return importer.ImportFromCSV(cmd.Context(), csvPath, opts)
	},
}

func init() {
	binder := authgearcmd.GetBinder()
	binder.BindString(
		cmdImport.PersistentFlags(),
		authgearcmd.ArgDatabaseSchema,
	)
	binder.BindString(
		cmdImport.PersistentFlags(),
		authgearcmd.ArgDatabaseURL,
	)
	binder.BindString(
		cmdImport.PersistentFlags(),
		authgearcmd.ArgAppID,
	)
	_ = cmdImport.Flags().Bool("email-case-sensitive", false, "set email is case sensitive")
	_ = cmdImport.Flags().Bool("email-block-plus-sign", false, "disallow plus sign in email")
	_ = cmdImport.Flags().Bool("email-ignore-dot-sign", false, "ignore the dot sign in email")
	_ = cmdImport.Flags().Bool("email-mark-as-verified", false, "mark email as verified")

	authgearcmd.Root.AddCommand(cmdImport)
}
