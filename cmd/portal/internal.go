package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var defaultAuthgearDomain string
var customAuthgearDomain string

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

		resourceDir := "./"
		if len(args) >= 1 {
			resourceDir = args[0]
		}

		internal.SetupPortal(&internal.SetupPortalOptions{
			DatabaseURL:           dbURL,
			DatabaseSchema:        dbSchema,
			DefaultAuthgearDoamin: defaultAuthgearDomain,
			CustomAuthgearDomain:  customAuthgearDomain,
			ResourceDir:           resourceDir,
		})

	},
}

func init() {
	cmdInternal.AddCommand(cmdInternalSetupPortal)

	cmdInternalSetupPortal.Flags().StringVar(&DatabaseURL, "database-url", "", "Database URL")
	cmdInternalSetupPortal.Flags().StringVar(&DatabaseSchema, "database-schema", "", "Database schema name")
	cmdInternalSetupPortal.Flags().StringVar(&defaultAuthgearDomain, "default-authgear-domain", "", "App default domain")
	cmdInternalSetupPortal.Flags().StringVar(&customAuthgearDomain, "custom-authgear-domain", "", "App custom domain")

	_ = cmdInternalSetupPortal.MarkFlagRequired("default-authgear-domain")
	_ = cmdInternalSetupPortal.MarkFlagRequired("custom-authgear-domain")
}
