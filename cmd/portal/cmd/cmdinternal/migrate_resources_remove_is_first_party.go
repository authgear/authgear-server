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

var cmdInternalMigrateRemoveIsFirstParty = &cobra.Command{
	Use:   "migrate-remove-is-first-party",
	Short: "Remove legacy is_first_party flag from the OAuth config",
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

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateRemoveIsFirstParty,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateRemoveIsFirstParty(appID string, configSourceData map[string]string, dryRun bool) error {
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

	oauthConfig, ok := m["oauth"].(map[string]interface{})
	if !ok {
		return nil
	}

	clients, ok := oauthConfig["clients"].([]interface{})
	if !ok {
		return fmt.Errorf("cannot read oauth.clients from authgear.yaml: %s", appID)
	}

	removedIsFirstParty := false
	for i, clientItf := range clients {
		client, ok := clientItf.(map[string]interface{})
		if !ok {
			return fmt.Errorf("cannot read oauth.clients[%d] from authgear.yaml: %s", i, appID)
		}
		if _, ok := client["is_first_party"]; ok {
			delete(client, "is_first_party")
			clients[i] = client
			removedIsFirstParty = true
		}
	}

	if !removedIsFirstParty {
		log.Printf("no legacy is_first_party flag, skip it: %s", appID)
		return nil
	}

	oauthConfig["clients"] = clients
	m["oauth"] = oauthConfig

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
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateRemoveIsFirstParty)
}
