package cmdinit

import (
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
}

var cmdInitConfig = &cobra.Command{
	Use:   "authgear.yaml",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		binder := authgearcmd.GetBinder()
		outputPath := binder.GetString(cmd, authgearcmd.ArgOutput)
		if outputPath == "" {
			outputPath = cmd.Use
		}

		opts := config.ReadAppConfigOptionsFromConsole()
		cfg := libconfig.GenerateAppConfigFromOptions(opts)
		err := config.MarshalConfigYAML(cfg, outputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

var cmdInitSecrets = &cobra.Command{
	Use:   "authgear.secrets.yaml",
	Short: "Initialize app secrets",
	Run: func(cmd *cobra.Command, args []string) {
		binder := authgearcmd.GetBinder()
		outputPath := binder.GetString(cmd, authgearcmd.ArgOutput)
		if outputPath == "" {
			outputPath = cmd.Use
		}

		opts := config.ReadSecretConfigOptionsFromConsole()
		createdAt := time.Now().UTC()
		cfg := libconfig.GenerateSecretConfigFromOptions(opts, createdAt, rand.SecureRand)
		err := config.MarshalConfigYAML(cfg, outputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

func init() {
	binder := authgearcmd.GetBinder()
	binder.BindString(cmdInit.PersistentFlags(), authgearcmd.ArgOutput)

	cmdInit.AddCommand(cmdInitConfig)
	cmdInit.AddCommand(cmdInitSecrets)

	authgearcmd.Root.AddCommand(cmdInit)
}
