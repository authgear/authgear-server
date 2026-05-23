package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EGenerateAppSessionToken)
	binder.BindString(cmdInternalE2EGenerateAppSessionToken.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EGenerateAppSessionToken.PersistentFlags(), ArgRefreshToken)
}

var cmdInternalE2EGenerateAppSessionToken = &cobra.Command{
	Use:   "generate-app-session-token",
	Short: "Exchange a refresh token for an app session token",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		refreshToken := binder.GetString(cmd, ArgRefreshToken)

		instance := e2e.End2End{}

		token, err := instance.GenerateAppSessionToken(cmd.Context(), appID, refreshToken)
		if err != nil {
			return err
		}

		fmt.Print(token)
		return nil
	},
}
