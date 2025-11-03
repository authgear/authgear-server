package cmd

import (
	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2ECreateSession)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgSessionType)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgSessionID)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgToken)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgClientID)
	binder.BindString(cmdInternalE2ECreateSession.PersistentFlags(), ArgSelectUserIDSQL)
}

var cmdInternalE2ECreateSession = &cobra.Command{
	Use:   "create-session",
	Short: "Create a session for a specific user",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		sessionType := binder.GetString(cmd, ArgSessionType)
		sessionID := binder.GetString(cmd, ArgSessionID)
		token := binder.GetString(cmd, ArgToken)
		clientID := binder.GetString(cmd, ArgClientID)
		selectUserIDSQL := binder.GetString(cmd, ArgSelectUserIDSQL)

		instance := e2e.End2End{}

		err := instance.CreateSession(
			cmd.Context(),
			appID,
			selectUserIDSQL,
			session.Type(sessionType),
			sessionID,
			clientID,
			token,
		)
		if err != nil {
			return err
		}

		return nil
	},
}
