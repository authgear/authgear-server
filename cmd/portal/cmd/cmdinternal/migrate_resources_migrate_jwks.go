package cmdinternal

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateJWKs = &cobra.Command{
	Use:   "migrate-jwks",
	Short: "Migrate jwks with use and algo",
	Run: func(cmd *cobra.Command, args []string) {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			log.Fatalf("%v", err.Error())
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			log.Fatalf("%v", err.Error())
		}

		internal.MigrateResources(cmd.Context(), &internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateResourcesJWKs,
			DryRun:                 &MigrateResourcesDryRun,
		})

	},
}

func ensureUseSigAndHS256Key(data map[interface{}]interface{}) map[interface{}]interface{} {
	keys := data["keys"].([]interface{})
	for i, key := range keys {
		perKey := key.(map[interface{}]interface{})
		perKey["use"] = "sig"
		perKey["alg"] = "HS256"
		keys[i] = perKey
	}
	data["keys"] = keys
	return data
}

func ensureUseSig(data map[interface{}]interface{}) map[interface{}]interface{} {
	keys := data["keys"].([]interface{})
	for i, key := range keys {
		perKey := key.(map[interface{}]interface{})
		perKey["use"] = "sig"
		keys[i] = perKey
	}
	data["keys"] = keys
	return data
}

func migrateResourcesJWKs(ctx context.Context, appID string, configSourceData map[string]string, DryRun bool) error {
	encodedData := configSourceData["authgear.secrets.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.secrets.yaml: %w", err)
	}

	if DryRun {
		log.Printf("Converting app secret (%s)", appID)
		log.Printf("Before updated:")
		log.Printf("\n%s\n", string(decoded))
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	secretsList := m["secrets"]
	secrets := secretsList.([]interface{})

	for i, s := range secrets {
		perSecret := s.(map[interface{}]interface{})
		data := perSecret["data"].(map[interface{}]interface{})
		key := perSecret["key"].(string)
		switch key {
		case "admin-api.auth", "oauth":
			newData := ensureUseSig(data)
			perSecret["data"] = newData
			secrets[i] = perSecret
		case "csrf", "webhook":
			newData := ensureUseSigAndHS256Key(data)
			perSecret["data"] = newData
			secrets[i] = perSecret
		default:
			continue
		}
	}

	secretsyaml, err := yaml.Marshal(&m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	if DryRun {
		log.Printf("After updated:")
		log.Printf("\n%s\n", string(secretsyaml))
	}

	configSourceData["authgear.secrets.yaml"] = base64.StdEncoding.EncodeToString(secretsyaml)
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateJWKs)
}
