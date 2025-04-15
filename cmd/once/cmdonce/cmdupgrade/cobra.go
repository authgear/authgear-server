package cmdupgrade

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

var CmdUpgrade = &cobra.Command{
	Use:           "upgrade",
	Short:         "Upgrade this program and Authgear",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err != nil {
				err = internal.PrintError(err)
			}
		}()

		return internal.ErrCommandUpgradeNotImplemented
	},
}
