package cmdimages

import (
	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
)

var CmdImages = &cobra.Command{
	Use:   "images",
	Short: "images commands",
}

func init() {
	authgearcmd.Root.AddCommand(CmdImages)
}
