package cmdinternal

import (
	"os"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/cmd/authgear/e2e"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/spf13/cobra"
)

func NewLoggerFactory() *log.Factory {
	return log.NewFactory(log.LevelInfo)
}

var cmdInternalE2EImportUser = &cobra.Command{
	Use:   "e2e import-users",
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
			ConfigSource: configsource.Config{
				Type:      configsource.TypeLocalFS,
				Directory: configSourceDir,
			},
		}

		err := instance.CreateUserFixtures(jsonPath)
		if err != nil {
			return err
		}

		return nil
	},
}
