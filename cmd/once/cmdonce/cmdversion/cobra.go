package cmdversion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

var CmdVersion = &cobra.Command{
	Use:           "version",
	Short:         "Show various versions about this program",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err != nil {
				err = internal.PrintError(err)
			}
		}()

		err = cobra.NoArgs(cmd, args)
		if err != nil {
			return
		}

		return
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%v %v\n", internal.ProgramName, internal.Version)
		fmt.Printf("license_server_endpoint %v\n", internal.GetLicenseServerEndpoint(cmd))
		fmt.Printf("image %v\n", internal.GetDockerImage(cmd))
		return nil
	},
}
