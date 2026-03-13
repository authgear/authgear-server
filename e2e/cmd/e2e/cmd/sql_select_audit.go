package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EQuerySQLSelectAudit)
	binder.BindString(cmdInternalE2EQuerySQLSelectAudit.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EQuerySQLSelectAudit.PersistentFlags(), ArgRawSQL)
}

var cmdInternalE2EQuerySQLSelectAudit = &cobra.Command{
	Use:   "query-sql-select-audit",
	Short: "Execute SQL SELECT queries on the audit database for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		rawSQL := binder.GetString(cmd, ArgRawSQL)

		instance := e2e.End2End{}

		dbRows, err := instance.QuerySQLSelectAudit(appID, rawSQL)
		if err != nil {
			return err
		}

		fmt.Print(dbRows)

		return nil
	},
}
