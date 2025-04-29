package cmdsetup

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
	"github.com/authgear/authgear-server/pkg/util/termutil"
)

var CmdSetup = &cobra.Command{
	Use:           "setup license-key",
	Short:         "Set up your Authgear ONCE installation.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err != nil {
				err = internal.PrintError(err)
			}
		}()

		err = cobra.ExactArgs(1)(cmd, args)
		if err != nil {
			return
		}

		// tea requires a terminal to run, so we do this checking prior to launching the tea program.
		if !termutil.IsTerminal(os.Stdout.Fd()) {
			err = fmt.Errorf("This command sets up your Authgear ONCE installation interactively. Thus, you must run it connected to a terminal.")
			return
		}

		return
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		image := internal.GetDockerImage(cmd)

		setupApp := SetupApp{Context: cmd.Context(), Image: image}
		prog := tea.NewProgram(setupApp)
		model, err := prog.Run()
		if err != nil {
			return err
		}

		setupApp = model.(SetupApp)
		if setupApp.HasError() {
			os.Exit(1)
		}
		return nil
	},
}
