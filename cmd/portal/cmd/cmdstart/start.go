//go:build !authgearlite
// +build !authgearlite

package cmdstart

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/server"
)

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}
		ctrl.Start(cmd.Context())
	},
}

func init() {
	portalcmd.Root.AddCommand(cmdStart)
}
