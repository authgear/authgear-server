package cmdinit

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	"github.com/authgear/authgear-server/cmd/authgear/config"
	libconfig "github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Run: func(cmd *cobra.Command, args []string) {
		binder := authgearcmd.GetBinder()
		outputFolderPath := binder.GetString(cmd, authgearcmd.ArgOutputFolder)
		if outputFolderPath == "" {
			outputFolderPath = "./"
		}

		// obtain options
		appConfigOpts := config.ReadAppConfigOptionsFromConsole()
		oauthClientConfigOpts, err := config.ReadOAuthClientConfigsFromConsole()
		if err != nil {
			log.Fatalf("invalid input: %s", err.Error())
			return
		}
		phoneOTPMode := config.ReadPhoneOTPMode()
		skipEmailVerification := config.ReadSkipEmailVerification()
		var appSecretsOpts *libconfig.GenerateSecretConfigOptions
		if forHelmChart, err := cmd.Flags().GetBool("for-helm-chart"); err == nil && forHelmChart {
			// Skip all the db, redis, elasticsearch credentials
			// Those are provided via the helm chart
			appSecretsOpts = &libconfig.GenerateSecretConfigOptions{}
		} else {
			appSecretsOpts = config.ReadSecretConfigOptionsFromConsole()
		}

		// generate app config
		appConfig := libconfig.GenerateAppConfigFromOptions(appConfigOpts)

		// generate oauth client for the portal
		oauthClientConfig, err := libconfig.GenerateOAuthConfigFromOptions(oauthClientConfigOpts)
		if err != nil {
			log.Fatalf("failed to generate oauth client config: %s", err.Error())
			return
		}

		// assign oauth client to app config
		if appConfig.OAuth == nil {
			appConfig.OAuth = &libconfig.OAuthConfig{}
		}
		appConfig.OAuth.Clients = append(appConfig.OAuth.Clients, *oauthClientConfig)

		// assign phone otp mode to app config
		appConfig.Authenticator = &libconfig.AuthenticatorConfig{
			OOB: &libconfig.AuthenticatorOOBConfig{
				SMS: &libconfig.AuthenticatorOOBSMSConfig{
					PhoneOTPMode: phoneOTPMode,
				},
			},
		}

		// assign email verification enabled
		emailVerificationEnabled := !skipEmailVerification
		appConfig.Verification = &libconfig.VerificationConfig{
			Claims: &libconfig.VerificationClaimsConfig{
				Email: &libconfig.VerificationClaimConfig{
					Enabled:  &emailVerificationEnabled,
					Required: &emailVerificationEnabled,
				},
			},
		}

		// generate secret config
		createdAt := time.Now().UTC()
		appSecretConfig := libconfig.GenerateSecretConfigFromOptions(appSecretsOpts, createdAt, rand.SecureRand)

		err = os.MkdirAll(outputFolderPath, 0755)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
			return
		}

		// write authgear.yaml
		err = config.MarshalConfigYAML(appConfig, outputFolderPath, "authgear.yaml")
		if err != nil {
			if errors.Is(err, config.ErrUserCancel) {
				return
			}
			log.Fatalf("cannot write authgear.yaml: %s", err.Error())
		}

		// write authgear.secrets.yaml
		err = config.MarshalConfigYAML(appSecretConfig, outputFolderPath, "authgear.secrets.yaml")
		if err != nil {
			if errors.Is(err, config.ErrUserCancel) {
				return
			}
			log.Fatalf("cannot write authgear.secrets.yaml: %s", err.Error())
		}
	},
}

func init() {
	binder := authgearcmd.GetBinder()

	binder.BindString(cmdInit.PersistentFlags(), authgearcmd.ArgOutputFolder)
	_ = cmdInit.Flags().Bool("for-helm-chart", false, "Generate config for helm chart deployment")

	authgearcmd.Root.AddCommand(cmdInit)
}
