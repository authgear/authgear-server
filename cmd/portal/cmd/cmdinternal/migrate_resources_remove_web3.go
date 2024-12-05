package cmdinternal

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateRemoveWeb3 = &cobra.Command{
	Use:   "migrate-remove-web3",
	Short: "Remove web3 related config",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		internal.MigrateResources(cmd.Context(), &internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateRemoveWeb3,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateRemoveWeb3(appID string, configSourceData map[string]string, dryRun bool) error {
	err := migrateRemoveWeb3_authgear_yaml(appID, configSourceData, dryRun)
	if err != nil {
		return err
	}

	err = migrateRemoveWeb3_authgear_features_yaml(appID, configSourceData, dryRun)
	if err != nil {
		return err
	}

	return nil
}

func migrateRemoveWeb3_authgear_yaml(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedAuthgearYAML, ok := configSourceData["authgear.yaml"]
	if !ok {
		return nil
	}

	decodedAuthgearYAML, err := base64.StdEncoding.DecodeString(encodedAuthgearYAML)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	authgearYAML := make(map[string]interface{})
	err = yaml.Unmarshal(decodedAuthgearYAML, &authgearYAML)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	changed := false

	// Delete web3 from authgear.yaml
	_, ok = authgearYAML["web3"]
	if ok {
		changed = true
		log.Printf("web3 is present in authgear.yaml: %v", appID)
		delete(authgearYAML, "web3")
	}

	// If authentication.identities contain siwe, then
	// authentication.identities = [login_id, oauth]
	// identity.login_id.keys = [{ type: "email" }]
	// authentication.primary_authenticators = ["password"]

	authentication, ok := authgearYAML["authentication"].(map[string]interface{})
	if ok {
		authenticationIdentitiesContainSIWE := false

		identities, ok := authentication["identities"].([]interface{})
		if ok {
			for _, i := range identities {
				if i, ok := i.(string); ok {
					if i == "siwe" {
						authenticationIdentitiesContainSIWE = true
						log.Printf("authentication.identities contain siwe: %v", appID)
					}
				}
			}
		}

		if authenticationIdentitiesContainSIWE {
			changed = true

			authentication["identities"] = []interface{}{
				"login_id",
				"oauth",
			}
			authentication["primary_authenticators"] = []interface{}{
				"password",
			}
			authgearYAML["identity"] = map[string]interface{}{
				"login_id": map[string]interface{}{
					"keys": []interface{}{
						map[string]interface{}{
							"type": "email",
						},
					},
				},
			}
		}
	}

	if changed {
		migrated, err := yaml.Marshal(authgearYAML)
		if err != nil {
			return fmt.Errorf("failed marshal yaml: %w", err)
		}
		configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	}

	return nil
}

func migrateRemoveWeb3_authgear_features_yaml(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedAuthgearFeaturesYAML, ok := configSourceData["authgear.features.yaml"]
	if !ok {
		return nil
	}

	decodedAuthgearFeaturesYAML, err := base64.StdEncoding.DecodeString(encodedAuthgearFeaturesYAML)
	if err != nil {
		return fmt.Errorf("failed decode authgear.features.yaml: %w", err)
	}

	authgearFeaturesYAML := make(map[string]interface{})
	err = yaml.Unmarshal(decodedAuthgearFeaturesYAML, &authgearFeaturesYAML)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	changed := false

	// Delete web3 from authgear.features.yaml
	_, ok = authgearFeaturesYAML["web3"]
	if ok {
		changed = true
		log.Printf("web3 is present in authgear.features.yaml: %v", appID)
		delete(authgearFeaturesYAML, "web3")
	}

	if changed {
		migrated, err := yaml.Marshal(authgearFeaturesYAML)
		if err != nil {
			return fmt.Errorf("failed marshal yaml: %w", err)
		}
		configSourceData["authgear.features.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	}

	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateRemoveWeb3)
}
