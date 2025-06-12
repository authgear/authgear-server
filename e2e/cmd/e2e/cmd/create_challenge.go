package cmd

import (
	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2ECreateChallenge)
	binder.BindString(cmdInternalE2ECreateChallenge.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2ECreateChallenge.PersistentFlags(), ArgToken)
	binder.BindString(cmdInternalE2ECreateChallenge.PersistentFlags(), ArgPurpose)
}

var cmdInternalE2ECreateChallenge = &cobra.Command{
	Use:   "create-challenge",
	Short: "Create a challenge",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		token := binder.GetString(cmd, ArgToken)
		purpose := binder.GetString(cmd, ArgPurpose)

		instance := e2e.End2End{}

		err := instance.CreateChallenge(
			cmd.Context(),
			appID,
			challenge.Purpose(purpose),
			token,
		)
		if err != nil {
			return err
		}

		return nil
	},
}
