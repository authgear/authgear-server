package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateExample = &cobra.Command{
	Use:   "example",
	Short: "Migrate resources example",
	Run: func(cmd *cobra.Command, args []string) {
		binder := getBinder()
		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			log.Fatalf(err.Error())
		}

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateResourcesExample,
			DryRun:                 &MigrateResourcesDryRun,
		})

	},
}

func migrateResourcesExample(appID string, configSourceData map[string]string, DryRun bool) error {
	// example update app accounts' dark_theme_disabled
	if appID != "accounts" {
		return nil
	}
	encodedData := configSourceData["authgear.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	log.Printf("Before updated:")
	log.Printf("\n%s\n", string(decoded))

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	uiSettingMap := m["ui"]
	var uiSetting map[interface{}]interface{}
	if uiSettingMap != nil {
		uiSetting = uiSettingMap.(map[interface{}]interface{})
	} else {
		uiSetting = make(map[interface{}]interface{})
	}

	uiSetting["dark_theme_disabled"] = true
	m["ui"] = uiSetting

	authgearyaml, err := yaml.Marshal(&m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	log.Printf("After updated:")
	log.Printf("\n%s\n", string(authgearyaml))

	configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(authgearyaml)
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateExample)
}
