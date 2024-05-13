package cmd

import (
	"os"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EImportUser)
	binder.BindString(cmdInternalE2EImportUser.PersistentFlags(), ArgAppID)
}

var cmdInternalE2EImportUser = &cobra.Command{
	Use:   "import-users [jsonPath]",
	Short: "Import users for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		jsonPath := args[0]

		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			return nil
		}

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.ImportUsers(appID, jsonPath)
		if err != nil {
			return err
		}

		return nil
	},
}
