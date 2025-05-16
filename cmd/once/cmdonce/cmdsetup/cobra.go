package cmdsetup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/once/cmdonce/internal"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/termutil"
)

func init() {
	_ = CmdSetup.Flags().Bool(
		"prompt-enable-certbot",
		internal.QuestionName_EnableCertbot_PromptByDefault,
		"Ask you whether to enable certbot.",
	)
	_ = CmdSetup.Flags().Bool(
		"prompt-certbot-environment",
		internal.QuestionName_SelectCertbotEnvironment_PromptByDefault,
		"Ask you which certbot environment to use",
	)
}

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
		if !termutil.StdinStdoutIsTerminal() {
			err = fmt.Errorf("This command sets up your Authgear ONCE installation interactively. Thus, you must run it connected to a terminal.")
			return
		}

		err = internal.CheckAllPublishedPortsNotListening()
		if err != nil {
			return
		}

		_, err = exec.LookPath(internal.BinDocker)
		if err != nil {
			err = errors.Join(err, internal.ErrNoDocker)
			return
		}

		return
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client := httputil.NewExternalClient(10 * time.Second)
		licenseKey := args[0]
		endpoint := internal.GetLicenseServerEndpoint(cmd)
		image := internal.GetDockerImage(cmd)

		volumeExists, err := internal.CheckVolumeExists(cmd.Context())
		if err != nil {
			err = internal.PrintError(err)
			return err
		}
		fingerprint := ""
		if volumeExists {
			fingerprint, err = internal.GetPersistentEnvironmentVariableInVolume(cmd.Context(), "AUTHGEAR_ONCE_MACHINE_FINGERPRINT")
			if err != nil {
				err = internal.PrintError(err)
				return err
			}
		} else {
			fingerprint = internal.GenerateMachineFingerprint()
		}

		licenseOpts := internal.LicenseOptions{
			Endpoint:    endpoint,
			LicenseKey:  licenseKey,
			Fingerprint: fingerprint,
		}

		_, err = internal.CheckLicense(ctx, client, licenseOpts)
		if err != nil {
			err = internal.PrintError(err)
			return err
		}

		if volumeExists {
			fmt.Printf("TODO: ResetupApp\n")
		} else {
			setupApp := SetupApp{
				Context:        ctx,
				HTTPClient:     client,
				LicenseOptions: licenseOpts,

				QuestionName_EnableCertbot_Prompt:            internal.FlagsGetBool(cmd, "prompt-enable-certbot"),
				QuestionName_SelectCertbotEnvironment_Prompt: internal.FlagsGetBool(cmd, "prompt-certbot-environment"),

				AUTHGEAR_ONCE_LICENSE_KEY:         licenseKey,
				AUTHGEAR_ONCE_MACHINE_FINGERPRINT: fingerprint,
				AUTHGEAR_ONCE_IMAGE:               image,
			}

			prog := tea.NewProgram(setupApp)
			model, err := prog.Run()
			if err != nil {
				return err
			}

			setupApp = model.(SetupApp)
			if setupApp.HasError() {
				os.Exit(1)
			}
		}

		return nil
	},
}
