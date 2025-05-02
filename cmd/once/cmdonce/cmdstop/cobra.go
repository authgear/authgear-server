package cmdstop

import (
	"os/exec"
	"slices"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

var CmdStop = &cobra.Command{
	Use:           "stop",
	Short:         "Stop Authgear",
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

		_, err = exec.LookPath(internal.BinDocker)
		if err != nil {
			err = internal.ErrNoDocker
		}

		return

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err != nil {
				err = internal.PrintError(err)
			}
		}()

		ctx := cmd.Context()
		cs, err := internal.DockerLs(ctx)
		if err != nil {
			return
		}
		ok := slices.ContainsFunc(cs, func(c internal.DockerContainer) bool {
			return c.Names == internal.NameDockerContainer
		})
		if !ok {
			err = internal.ErrDockerContainerNotExists
			return
		}
		err = internal.DockerStop(ctx, internal.NameDockerContainer)
		if err != nil {
			return
		}

		return
	},
}
