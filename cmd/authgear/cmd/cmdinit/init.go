package cmdinit

import (
	"errors"
	"log"
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

		appSecretsOpts := config.ReadSecretConfigOptionsFromConsole()

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

		// generate secret config
		createdAt := time.Now().UTC()
		appSecretConfig := libconfig.GenerateSecretConfigFromOptions(appSecretsOpts, createdAt, rand.SecureRand)

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
	authgearcmd.Root.AddCommand(cmdInit)
}
