package cmdonce

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdsetup"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdstart"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdstop"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdupgrade"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdversion"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

func init() {
	CmdRoot.AddCommand(cmdversion.CmdVersion)
	CmdRoot.AddCommand(cmdsetup.CmdSetup)
	CmdRoot.AddCommand(cmdstart.CmdStart)
	CmdRoot.AddCommand(cmdstop.CmdStop)
	CmdRoot.AddCommand(cmdupgrade.CmdUpgrade)

	_ = CmdRoot.PersistentFlags().String("image", "", "Override the default image")
	_ = CmdRoot.PersistentFlags().MarkHidden("image")

	if internal.LicenseServerEndpointOverridable {
		_ = CmdRoot.PersistentFlags().String("license-server-endpoint", "", "Override the license server endpoint")
		_ = CmdRoot.PersistentFlags().MarkHidden("license-server-endpoint")
	}
}

var CmdRoot = &cobra.Command{
	Use: internal.ProgramName,
	// Suppress the --version flag.
	Version: "",
	CompletionOptions: cobra.CompletionOptions{
		// Hide the subcommand `completion`, instead of disabling it.
		HiddenDefaultCmd: true,
	},
}

func Run(ctx context.Context) {
	err := CmdRoot.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
