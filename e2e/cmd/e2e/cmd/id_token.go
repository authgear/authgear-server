package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EGenerateIDToken)
	binder.BindString(cmdInternalE2EGenerateIDToken.PersistentFlags(), ArgAppID)
}

var cmdInternalE2EGenerateIDToken = &cobra.Command{
	Use:   "generate-id-token [user_id]",
	Short: "Generate ID Token with user id",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		userID := args[0]

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		idToken, err := instance.GenerateIDToken(appID, userID)
		if err != nil {
			return err
		}

		fmt.Print(idToken)

		return nil
	},
}
