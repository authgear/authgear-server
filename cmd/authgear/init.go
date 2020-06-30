package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/skygeario/skygear-server/cmd/authgear/config"
)

var InitConfigOutputPath string
var InitSecretsOutputPath string

var cmdInit = &cobra.Command{
	Use:   "init [config|secret]",
	Short: "Initialize configuration",
}

var cmdInitConfig = &cobra.Command{
	Use:   "config",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		opts := config.ReadOptionsFromConsole()
		cfg := config.NewAppConfigFromOptions(opts)
		err := config.MarshalConfigYAML(cfg, InitConfigOutputPath)
		if err != nil {
			log.Fatalf("cannot write file: %s", err.Error())
		}
	},
}

var cmdInitSecrets = &cobra.Command{
	Use:   "secrets",
	Short: "Initialize app secrets",
	Run: func(cmd *cobra.Command, args []string) {
		opts := config.ReadSecretOptionsFromConsole()
		cfg := config.NewSecretConfigFromOptions(opts)
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
