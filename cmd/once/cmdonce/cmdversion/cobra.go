package cmdversion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

var CmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show various versions about this program",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%v %v\n", internal.ProgramName, internal.Version)
		fmt.Printf("license_server_endpoint %v\n", internal.GetLicenseServerEndpoint(cmd))
		fmt.Printf("image %v\n", internal.GetDockerImage(cmd))
		return nil
	},
}
