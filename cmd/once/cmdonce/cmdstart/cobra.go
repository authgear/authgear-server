package cmdstart

import (
	"os/exec"
	"slices"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
)

var CmdStart = &cobra.Command{
	Use:           "start",
	Short:         "Start Authgear",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if err != nil {
				err = internal.PrintError(err)
			}
		}()

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

		image, err := cmd.Flags().GetString("image")
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		volumes, err := internal.DockerVolumeLs(ctx)
		if err != nil {
			return
		}
		volumeExists := slices.ContainsFunc(volumes, func(v internal.DockerVolume) bool {
			return v.Name == internal.NameDockerVolume && v.Scope == internal.DockerVolumeScopeLocal
		})

		cs, err := internal.DockerLs(ctx)
		if err != nil {
			return
		}
		containerExists := slices.ContainsFunc(cs, func(c internal.DockerContainer) bool {
			return c.Names == internal.NameDockerContainer
		})

		if !volumeExists {
			err = internal.ErrDockerContainerNotExists
			return
		}

		if !containerExists {
			// Run the container without providing any environment variables.
			// We assume the environment variables are persisted in the volume.
			opts := internal.NewDockerRunOptionsForStarting(image)
			err = internal.DockerRun(ctx, opts)
			if err != nil {
				return
			}
		}

		err = internal.DockerStart(ctx, internal.NameDockerContainer)
		if err != nil {
			return
		}

		return
	},
}
