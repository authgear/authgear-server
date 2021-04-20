package main

import (
	"crypto/rand"
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/config"
	libconfig "github.com/authgear/authgear-server/pkg/lib/config"
)

var InitConfigOutputPath string
var InitSecretsOutputPath string

var cmdInit = &cobra.Command{
	Use:   "init [authgear.yaml|authgear.secrets.yaml]",
	Short: "Initialize configuration",
}

var cmdInitConfig = &cobra.Command{
	Use:   "authgear.yaml",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		opts := config.ReadAppConfigOptionsFromConsole()
		cfg := libconfig.GenerateAppConfigFromOptions(opts)
		err := config.MarshalConfigYAML(cfg, InitConfigOutputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

var cmdInitSecrets = &cobra.Command{
	Use:   "authgear.secrets.yaml",
	Short: "Initialize app secrets",
	Run: func(cmd *cobra.Command, args []string) {
		opts := config.ReadSecretConfigOptionsFromConsole()
		cfg := libconfig.GenerateSecretConfigFromOptions(opts, rand.Reader)
		err := config.MarshalConfigYAML(cfg, InitSecretsOutputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

func init() {
	cmdInit.AddCommand(cmdInitConfig)
	cmdInit.AddCommand(cmdInitSecrets)

	cmdInitConfig.Flags().StringVarP(&InitConfigOutputPath, "output", "o", "authgear.yaml", "Output YAML path")
	cmdInitSecrets.Flags().StringVarP(&InitSecretsOutputPath, "output", "o", "authgear.secrets.yaml", "Output YAML path")
}
