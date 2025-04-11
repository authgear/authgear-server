package cmdsetup

import (
	"fmt"
	"os"
	"os/exec"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
	"github.com/authgear/authgear-server/pkg/util/termutil"
)

var CmdSetup = &cobra.Command{
	Use:          "setup",
	Short:        "Set up your Authgear ONCE installation.",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !termutil.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf("This command sets up your Authgear ONCE installation interactively. Thus, you must run it connected to a terminal.")
		}

		_, err := exec.LookPath(internal.BinDocker)
		if err != nil {
			return fmt.Errorf("%v is not installed on your machine. Please install it first. See https://docs.docker.com/get-started/get-docker/", internal.BinDocker)
		}

		volumes, err := internal.DockerVolumeLs(cmd.Context())
		if err != nil {
			return err
		}

		if slices.ContainsFunc(volumes, func(v internal.DockerVolume) bool {
			return v.Name == internal.NameDockerVolume && v.Scope == internal.DockerVolumeScopeLocal
		}) {
			return fmt.Errorf(
				"The docker volume %v exists already.\nHave you set up this before?\nIf yes, then you should the following to start Authgear\n\n  %v start\n",
				internal.NameDockerVolume,
				internal.ProgramName,
			)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		setupApp := SetupApp{}
		prog := tea.NewProgram(setupApp)
		model, err := prog.Run()
		if err != nil {
			return err
		}

		setupApp = model.(SetupApp)
		// The user quits the setup. Just exit 0.
		if !setupApp.Complete {
			return nil
		}

		_ = setupApp.ToResult()
		return nil
	},
}
