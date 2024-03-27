package cmdinternal

import (
	"os"
	"path/filepath"

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

var cmdInternalE2ECreateConfigSource = &cobra.Command{
	Use:   "create-configsource",
	Short: "Create a config source record in the database with the given config source directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		configSourceDir := binder.GetString(cmd, authgearcmd.ArgConfigSourceDir)
		appID := filepath.Base(configSourceDir)

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.CreateConfigSource(appID, configSourceDir)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalE2EImportUser = &cobra.Command{
	Use:   "import-users",
	Short: "Import users for e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		configSourceDir := binder.GetString(cmd, authgearcmd.ArgConfigSourceDir)
		jsonPath := configSourceDir + "/users.json"

		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			return nil
		}

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.ImportUsers(configSourceDir, jsonPath)
		if err != nil {
			return err
		}

		return nil
	},
}
