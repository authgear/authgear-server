package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var defaultDomain string
var customDomain string

var cmdInternal = &cobra.Command{
	Use:   "internal [setup-portal]",
	Short: "Setup portal config source data in db",
}

var cmdInternalSetupPortal = &cobra.Command{
	Use:   "setup-portal",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			log.Fatalf("failed to create app: %s", err)
		}

		internal.SetupPortal(&internal.SetupPortalOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			DefaultDoamin:  defaultDomain,
			CustomDomain:   customDomain,
		})

	},
}

func init() {
	cmdInternal.AddCommand(cmdInternalSetupPortal)

	cmdInternalSetupPortal.Flags().StringVar(&DatabaseURL, "database-url", "", "Database URL")
	cmdInternalSetupPortal.Flags().StringVar(&DatabaseSchema, "database-schema", "", "Database schema name")
	cmdInternalSetupPortal.Flags().StringVar(&defaultDomain, "default-domain", "", "App default domain")
	cmdInternalSetupPortal.Flags().StringVar(&customDomain, "custom-domain", "", "App custom domain")

	_ = cmdInternalSetupPortal.MarkFlagRequired("default-domain")
	_ = cmdInternalSetupPortal.MarkFlagRequired("custom-domain")
}
