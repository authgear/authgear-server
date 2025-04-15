package cmdonce

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdsetup"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdstart"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/cmdstop"
	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

func init() {
	CmdRoot.AddCommand(cmdsetup.CmdSetup)
	CmdRoot.AddCommand(cmdstart.CmdStart)
	CmdRoot.AddCommand(cmdstop.CmdStop)

	_ = CmdRoot.PersistentFlags().String("image", "", "Override the default image")
	_ = CmdRoot.PersistentFlags().MarkHidden("image")
}

var CmdRoot = &cobra.Command{
	Use:     internal.ProgramName,
	Version: internal.Version,
}

func Run(ctx context.Context) {
	err := CmdRoot.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
