package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

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
		marshalYAML(cfg, InitConfigOutputPath)
	},
}

var cmdInitSecrets = &cobra.Command{
	Use:   "secrets",
	Short: "Initialize app secrets",
	Run: func(cmd *cobra.Command, args []string) {
		opts := config.ReadSecretOptionsFromConsole()
		cfg := config.NewSecretConfigFromOptions(opts)
		marshalYAML(cfg, InitSecretsOutputPath)
	},
}

func marshalYAML(cfg interface{}, output string) {
	yaml, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	if output == "-" {
		_, err = os.Stdout.Write(yaml)
	} else {
		err = ioutil.WriteFile(output, yaml, os.ModePerm)
	}
	if err != nil {
		log.Fatalf("cannot write config: %s", err.Error())
	}
}

func init() {
	cmdInit.AddCommand(cmdInitConfig)
	cmdInit.AddCommand(cmdInitSecrets)

	cmdInitConfig.Flags().StringVarP(&InitConfigOutputPath, "output", "o", "authgear.yaml", "Output YAML path")
	cmdInitSecrets.Flags().StringVarP(&InitSecretsOutputPath, "output", "o", "authgear.secrets.yaml", "Output YAML path")
}
