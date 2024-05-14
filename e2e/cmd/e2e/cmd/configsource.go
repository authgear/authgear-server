package cmd

import (
	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2ECreateConfigSource)
	binder.BindString(cmdInternalE2ECreateConfigSource.PersistentFlags(), ArgAppID)
	binder.BindString(cmdInternalE2ECreateConfigSource.PersistentFlags(), ArgConfigSource)
	binder.BindString(cmdInternalE2ECreateConfigSource.PersistentFlags(), ArgConfigOverride)
}

var cmdInternalE2ECreateConfigSource = &cobra.Command{
	Use:   "create-configsource",
	Short: "Create a config source record in the database with the given config source directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		configSource := binder.GetString(cmd, ArgConfigSource)
		configOverride := binder.GetString(cmd, ArgConfigOverride)

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		err := instance.CreateApp(appID, configSource, configOverride)
		if err != nil {
			return err
		}

		return nil
	},
}
