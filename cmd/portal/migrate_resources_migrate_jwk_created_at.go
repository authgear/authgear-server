package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateJWKCreatedAt = &cobra.Command{
	Use:   "migrate-jwk-created-at",
	Short: "Add created_at to the JWK if not exist",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()

		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateJWKCreatedAt,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateJWKCreatedAt(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedData := configSourceData["authgear.secrets.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.secrets.yaml: %w", err)
	}

	if dryRun {
		log.Printf("Converting app (%s)", appID)
		log.Printf("Before updated:")
		log.Printf("\n%s\n", string(decoded))
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	createdAt := float64(time.Now().UTC().Unix())

	secrets := m["secrets"].([]interface{})
	for idx, secretItemIface := range secrets {
		secretItem := secretItemIface.(map[string]interface{})
		key := secretItem["key"].(string)
		if key == "csrf" || key == "webhook" || key == "oauth" || key == "admin-api.auth" {
			data := secretItem["data"].(map[string]interface{})
			dataBytes, err := json.Marshal(data)
			if err != nil {
				return err
			}
			jwkSet, err := jwk.Parse(dataBytes)
			if err != nil {
				return err
			}
			for i := 0; i < jwkSet.Len(); i++ {
				jwkKey, ok := jwkSet.Get(i)
				if ok {
					_, ok := jwkKey.Get("created_at")
					if !ok {
						_ = jwkKey.Set("created_at", createdAt)
					}
				}
			}

			dataBytes, err = json.Marshal(jwkSet)
			if err != nil {
				return err
			}

			var dataJSON map[string]interface{}
			err = json.Unmarshal(dataBytes, &dataJSON)
			if err != nil {
				return err
			}

			secretItem["data"] = dataJSON

			secrets[idx] = secretItem
		}
	}

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	if dryRun {
		log.Printf("After updated:")
		log.Printf("\n%s\n", string(migrated))
	}

	configSourceData["authgear.secrets.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateJWKCreatedAt)
}
