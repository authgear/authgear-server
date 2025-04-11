package cmdsetup

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/termutil"
)

var CmdSetup = &cobra.Command{
	Use:          "setup",
	Short:        "Set up your Authgear ONCE installation.",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// tea requires a terminal to run, so we do this checking prior to launching the tea program.
		if !termutil.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf("This command sets up your Authgear ONCE installation interactively. Thus, you must run it connected to a terminal.")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		setupApp := SetupApp{Context: cmd.Context()}
		prog := tea.NewProgram(setupApp)
		model, err := prog.Run()
		if err != nil {
			return err
		}

		setupApp = model.(SetupApp)
		if setupApp.FatalErr != nil {
			os.Exit(1)
		}
		return nil
	},
}
