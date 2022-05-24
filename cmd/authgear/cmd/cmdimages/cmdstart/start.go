package cmdstart

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/cmd/cmdimages"
	"github.com/authgear/authgear-server/cmd/authgear/images/server"
)

var cmdImagesStart = &cobra.Command{
	Use:   "start",
	Short: "Start images server",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}
		ctrl.Start()
	},
}

func init() {
	cmdimages.CmdImages.AddCommand(cmdImagesStart)
}
