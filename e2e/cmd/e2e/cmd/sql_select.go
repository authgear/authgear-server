package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EQuerySQLSelect)
	binder.BindString(cmdInternalE2EQuerySQLSelect.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2EQuerySQLSelect.PersistentFlags(), ArgRawSQL)
}

var cmdInternalE2EQuerySQLSelect = &cobra.Command{
	Use:   "query-sql-select",
	Short: "Execute SQL SELECT queries for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		rawSQL := binder.GetString(cmd, ArgRawSQL)

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		dbRows, err := instance.QuerySQLSelect(appID, rawSQL)
		if err != nil {
			return err
		}

		fmt.Print(dbRows)

		return nil
	},
}
