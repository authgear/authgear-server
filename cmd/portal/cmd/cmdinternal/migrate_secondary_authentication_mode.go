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

var cmdInternalMigrateSecondaryAuthenticationMode = &cobra.Command{
	Use:   "migrate-secondary-authentication-mode",
	Short: "Migrate secondary_authentication_mode",
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
			UpdateConfigSourceFunc: migrateSecondaryAuthenticationMode,
			DryRun:                 &MigrateResourcesDryRun,
		})
	},
}

func migrateSecondaryAuthenticationMode(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedData := configSourceData["authgear.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
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

	authentication, ok := m["authentication"].(map[string]interface{})
	if !ok {
		return nil
	}

	updated := false

	if mode, ok := authentication["secondary_authentication_mode"].(string); ok {
		// Turn "if_requested" into "disabled"
		if mode == "if_requested" {
			authentication["secondary_authentication_mode"] = "disabled"
			updated = true
		}
	}
	if secondaryAuthenticators, ok := authentication["secondary_authenticators"].([]interface{}); ok {
		// Treat empty secondary_authenticators as disabled.
		if secondaryAuthenticators != nil && len(secondaryAuthenticators) == 0 {
			authentication["secondary_authentication_mode"] = "disabled"
			delete(authentication, "secondary_authenticators")
			updated = true
		}
	}

	if !updated {
		return nil
	}

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	if dryRun {
		log.Printf("After updated:")
		log.Printf("\n%s\n", string(migrated))
	}

	configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateSecondaryAuthenticationMode)
}
