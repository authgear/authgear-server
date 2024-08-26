package cmdinternal

import (
	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
)

var cmdInternal = &cobra.Command{
	Use:   "internal",
	Short: "Internal commands",
}

func init() {
	authgearcmd.Root.AddCommand(cmdInternal)
}
