package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalMigrateExample = &cobra.Command{
	Use:   "example",
	Short: "Migrate resources example",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("missing db config: %s", err)
		}

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateResourcesExample,
			DryRun:                 &MigrateResourcesDryRun,
		})

	},
}

func migrateResourcesExample(appID string, configSourceData map[string]string) error {
	// FIXME: more concrete example
	if appID == "accounts" {
		configSourceData["portal"] = "true"
	}
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateExample)
}
