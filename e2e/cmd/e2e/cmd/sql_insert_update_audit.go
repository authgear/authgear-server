package cmd

import (
	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EExecuteSQLInsertUpdateAudit)
	binder.BindString(cmdInternalE2EExecuteSQLInsertUpdateAudit.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EExecuteSQLInsertUpdateAudit.PersistentFlags(), ArgCustomSQL)
}

var cmdInternalE2EExecuteSQLInsertUpdateAudit = &cobra.Command{
	Use:   "exec-sql-insert-update-audit",
	Short: "Execute custom SQL INSERT/UPDATE on the audit database for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		customSQL := binder.GetString(cmd, ArgCustomSQL)

		instance := e2e.End2End{}

		err := instance.ExecuteSQLInsertUpdateAudit(appID, customSQL)
		if err != nil {
			return err
		}

		return nil
	},
}
