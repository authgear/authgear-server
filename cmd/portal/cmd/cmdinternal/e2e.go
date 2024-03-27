package cmdinternal

import (
	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/e2e"
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
		binder := portalcmd.GetBinder()

		configSourceDir := binder.GetString(cmd, portalcmd.ArgConfigSourceDir)
		appID := binder.GetString(cmd, portalcmd.ArgAppID)

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.CreateApp(appID, configSourceDir)
		if err != nil {
			return err
		}

		return nil
	},
}
