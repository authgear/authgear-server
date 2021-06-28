package main

import (
	"crypto/rand"
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/config"
	libconfig "github.com/authgear/authgear-server/pkg/lib/config"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
}

var cmdInitConfig = &cobra.Command{
	Use:   "authgear.yaml",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		binder := getBinder()
		outputPath := binder.GetString(cmd, ArgOutput)
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
		binder := getBinder()
		outputPath := binder.GetString(cmd, ArgOutput)
		if outputPath == "" {
			outputPath = cmd.Use
		}

		opts := config.ReadSecretConfigOptionsFromConsole()
		cfg := libconfig.GenerateSecretConfigFromOptions(opts, rand.Reader)
		err := config.MarshalConfigYAML(cfg, outputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

func init() {
	binder := getBinder()
	binder.BindString(cmdInit.PersistentFlags(), ArgOutput)

	cmdInit.AddCommand(cmdInitConfig)
	cmdInit.AddCommand(cmdInitSecrets)
}
