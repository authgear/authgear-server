//go:build !authgearlite
// +build !authgearlite

package cmdstart

import (
	"log"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/server"
)

var cmdStart = &cobra.Command{
	Use:   "start [portal|siteadmin]...",
	Short: "Start specified servers",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}

		serverTypes := args
		if len(serverTypes) == 0 {
			// Default to start portal only
			serverTypes = []string{"portal"}
		}
		for _, typ := range serverTypes {
			switch typ {
			case "portal":
				ctrl.ServePortal = true
			case "siteadmin":
				ctrl.ServeSiteadmin = true
			default:
				log.Fatalf("unknown server type: %s", typ)
			}
		}

		ctrl.Start(cmd.Context())
	},
}

func init() {
	portalcmd.Root.AddCommand(cmdStart)
}
