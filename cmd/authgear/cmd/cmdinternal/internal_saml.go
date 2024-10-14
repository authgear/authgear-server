package cmdinternal

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	cmdInternal.AddCommand(cmdInternalSaml)
	cmdInternalSaml.AddCommand(cmdInternalSamlGenerateSigningKey)
}

var cmdInternalSaml = &cobra.Command{
	Use:   "saml",
	Short: "SAML commands",
}

var cmdInternalSamlGenerateSigningKey = &cobra.Command{
	Use:   "generate-signing-key { common-name }",
	Short: "Generate a signing key with a X.509 certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("common-name is required")
		}
		commonName := args[0]

		signingSecret, err := config.GenerateSAMLIdpSigningCertificate(commonName)
		if err != nil {
			return err
		}

		jsonBytes, err := json.Marshal(*signingSecret)
		if err != nil {
			return err
		}

		yamlBytes, err := yaml.JSONToYAML(jsonBytes)
		if err != nil {
			return err
		}

		fmt.Println(string(yamlBytes))
		return nil
	},
}
