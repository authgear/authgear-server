package cmd

import (
	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EExecuteCustomSQL)
	binder.BindString(cmdInternalE2EExecuteCustomSQL.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EExecuteCustomSQL.PersistentFlags(), ArgCustomSQL)
}

var cmdInternalE2EExecuteCustomSQL = &cobra.Command{
	Use:   "exec-sql",
	Short: "Execute custom SQL for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		customSQL := binder.GetString(cmd, ArgCustomSQL)

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.ExecuteCustomSQL(appID, customSQL)
		if err != nil {
			return err
		}

		return nil
	},
}
