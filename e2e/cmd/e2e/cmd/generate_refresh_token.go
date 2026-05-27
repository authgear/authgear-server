package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EGenerateRefreshToken)
	binder.BindString(cmdInternalE2EGenerateRefreshToken.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EGenerateRefreshToken.PersistentFlags(), ArgUserID)
	binder.BindString(cmdInternalE2EGenerateRefreshToken.PersistentFlags(), ArgClientID)
}

var cmdInternalE2EGenerateRefreshToken = &cobra.Command{
	Use:   "generate-refresh-token",
	Short: "Create an offline grant and return an encoded refresh token for the given user",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		userID := binder.GetString(cmd, ArgUserID)
		clientID := binder.GetString(cmd, ArgClientID)

		instance := e2e.End2End{}

		token, err := instance.GenerateRefreshToken(cmd.Context(), appID, userID, clientID)
		if err != nil {
			return err
		}

		fmt.Print(token)
		return nil
	},
}
