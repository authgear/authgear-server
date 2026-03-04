//go:build !authgearlite
// +build !authgearlite

package cmdstart

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/server"
)

var cmdStart = &cobra.Command{
	Use:   "start [portal] [superadmin]",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		// Default: start portal only
		ctrl := &server.Controller{
			PortalMode:     true,
			SuperadminMode: false,
		}

		// If any args provided, reset defaults and parse them
		if len(args) > 0 {
			ctrl.PortalMode = false
			ctrl.SuperadminMode = false

			// Check each argument
			for _, arg := range args {
				switch arg {
				case "portal":
					ctrl.PortalMode = true
				case "superadmin":
					ctrl.SuperadminMode = true
				}
			}

			// If no valid modes were recognized, default to portal
			if !ctrl.PortalMode && !ctrl.SuperadminMode {
				ctrl.PortalMode = true
			}
		}

		ctrl.Start(cmd.Context())
	},
}

func init() {
	portalcmd.Root.AddCommand(cmdStart)
}
