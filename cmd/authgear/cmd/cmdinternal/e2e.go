package cmdinternal

import (
	"os"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/cmd/authgear/e2e"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/spf13/cobra"
)

func NewLoggerFactory() *log.Factory {
	return log.NewFactory(log.LevelInfo)
}

var cmdInternalE2E = &cobra.Command{
	Use:   "e2e",
	Short: "End2End commands",
}

var cmdInternalE2EImportUser = &cobra.Command{
	Use:   "import-users [jsonPath]",
	Short: "Import users for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		appID := binder.GetString(cmd, authgearcmd.ArgAppID)
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

var cmdInternalE2EExecuteCustomSQL = &cobra.Command{
	Use:   "exec-sql",
	Short: "Execute custom SQL for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		appID := binder.GetString(cmd, authgearcmd.ArgAppID)
		customSQL := binder.GetString(cmd, authgearcmd.ArgCustomSQL)

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
