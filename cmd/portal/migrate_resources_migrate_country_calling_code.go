package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateCountryCallingCode = &cobra.Command{
	Use:   "migrate-country-calling-code",
	Short: "Remove legacy country_calling_code config",
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
			UpdateConfigSourceFunc: migrateCountryCallingCode,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateCountryCallingCode(appID string, configSourceData map[string]string, dryRun bool) error {
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

	ui, ok := m["ui"].(map[string]interface{})
	if !ok {
		return nil
	}

	_, ok = ui["country_calling_code"]
	if !ok {
		return nil
	}

	delete(ui, "country_calling_code")
	if len(ui) == 0 {
		delete(m, "ui")
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
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateCountryCallingCode)
}
