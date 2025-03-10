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
	"github.com/authgear/authgear-server/pkg/util/cliutil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		binder := authgearcmd.GetBinder()
		outputFolderPath := binder.GetString(cmd, authgearcmd.ArgOutputFolder)
		if outputFolderPath == "" {
			outputFolderPath = "./"
		}

		// obtain options
		appConfigOpts, err := config.ReadAppConfigOptionsFromConsole(ctx, cmd)
		if err != nil {
			return err
		}

		oauthClientConfigOpts, err := config.ReadOAuthClientConfigsFromConsole(ctx, cmd)
		if err != nil {
			return err
		}

		phoneOTPMode, err := config.ReadPhoneOTPMode(ctx, cmd)
		if err != nil {
			return err
		}

		skipEmailVerification, err := config.ReadSkipEmailVerification(ctx, cmd)
		if err != nil {
			return err
		}

		searchImpl, err := config.ReadSearchImplementation(ctx, cmd)
		if err != nil {
			return err
		}

		var appSecretsOpts *libconfig.GenerateSecretConfigOptions
		if forHelmChart, err := cmd.Flags().GetBool("for-helm-chart"); err == nil && forHelmChart {
			// Skip all the db, redis, elasticsearch credentials
			// Those are provided via the helm chart
			appSecretsOpts = &libconfig.GenerateSecretConfigOptions{}
		} else {
			appSecretsOpts, err = config.ReadSecretConfigOptionsFromConsole(ctx, cmd, searchImpl)
			if err != nil {
				return err
			}
		}

		// generate app config
		appConfig := libconfig.GenerateAppConfigFromOptions(appConfigOpts)

		// generate oauth client for the portal
		oauthClientConfig, err := libconfig.GenerateOAuthConfigFromOptions(oauthClientConfigOpts)
		if err != nil {
			return err
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

		// Set search implementation
		appConfig.Search = &libconfig.SearchConfig{
			Implementation: searchImpl,
		}

		// generate secret config
		createdAt := time.Now().UTC()
		appSecretConfig := libconfig.GenerateSecretConfigFromOptions(appSecretsOpts, createdAt, rand.SecureRand)

		err = os.MkdirAll(outputFolderPath, 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}

		// write authgear.yaml
		err = config.MarshalConfigYAML(ctx, cmd, appConfig, outputFolderPath, "authgear.yaml")
		if err != nil {
			if errors.Is(err, config.ErrUserCancel) {
				return nil
			}
			log.Fatalf("cannot write authgear.yaml: %s", err.Error())
		}

		// write authgear.secrets.yaml
		err = config.MarshalConfigYAML(ctx, cmd, appSecretConfig, outputFolderPath, "authgear.secrets.yaml")
		if err != nil {
			if errors.Is(err, config.ErrUserCancel) {
				return nil
			}
			log.Fatalf("cannot write authgear.secrets.yaml: %s", err.Error())
		}

		return nil
	},
}

func init() {
	binder := authgearcmd.GetBinder()

	binder.BindString(cmdInit.PersistentFlags(), authgearcmd.ArgOutputFolder)
	_ = cmdInit.Flags().Bool("for-helm-chart", false, "Generate config for helm chart deployment")
	_ = cmdInit.Flags().Bool("overwrite", false, "overwrite files if they exist already")

	cliutil.DefineFlagInteractive(cmdInit)
	config.Prompt_AppID.DefineFlag(cmdInit)
	config.Prompt_PublicOrigin.DefineFlag(cmdInit)
	config.Prompt_PortalOrigin.DefineFlag(cmdInit)
	config.Prompt_PortalClientID.DefineFlag(cmdInit)
	config.Prompt_PhoneOTPMode.DefineFlag(cmdInit)
	config.Prompt_DisableEmailVerification.DefineFlag(cmdInit)
	config.Prompt_SearchImplementation.DefineFlag(cmdInit)
	config.Prompt_DatabaseURL.DefineFlag(cmdInit)
	config.Prompt_DatabaseSchema.DefineFlag(cmdInit)
	config.Prompt_AuditDatabaseURL.DefineFlag(cmdInit)
	config.Prompt_AuditDatabaseSchema.DefineFlag(cmdInit)
	config.Prompt_SearchDatabaseURL.DefineFlag(cmdInit)
	config.Prompt_SearchDatabaseSchema.DefineFlag(cmdInit)
	config.Prompt_ElasticsearchURL.DefineFlag(cmdInit)
	config.Prompt_RedisURL.DefineFlag(cmdInit)
	config.Prompt_AnalyticRedisURL.DefineFlag(cmdInit)

	authgearcmd.Root.AddCommand(cmdInit)
}
