package cmdinternal

import (
	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/spf13/cobra"
)

var cmdInternalCheck = &cobra.Command{
	Use:   "check",
	Short: "Check integrity of data",
}

var cmdInternalCheckConfigSources = &cobra.Command{
	Use:   "config-sources",
	Short: "Check integrity of config sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		err = internal.CheckConfigSources(&internal.CheckConfigSourcesOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			AppIDs:         args,
		})

		return err
	},
}
