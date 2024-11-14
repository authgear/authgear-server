package cmdbackground

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/background"
	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
)

var cmdBackground = &cobra.Command{
	Use:   "background",
	Short: "Start the background job runner",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &background.Controller{}
		ctrl.Start(cmd.Context())
	},
}

func init() {
	authgearcmd.Root.AddCommand(cmdBackground)
}
