package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var defaultAuthgearDomain string
var customAuthgearDomain string
var kubeConfigPath string
var namespace string

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

var cmdInternalBreakingChange = &cobra.Command{
	Use:   "breaking-change",
	Short: "Commands for dealing with breaking changes",
}

var cmdInternalBreakingChangeMigrateK8SToDB = &cobra.Command{
	Use:   "migrate-k8s-to-db",
	Short: "Migrate config source from Kubernetes to database",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL, dbSchema, err := loadDBCredentials()
		if err != nil {
			return err
		}

		err = internal.MigrateK8SToDB(&internal.MigrateK8SToDBOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			KubeConfigPath: kubeConfigPath,
			Namespace:      namespace,
		})

		return err
	},
}

func init() {
	cmdInternal.AddCommand(cmdInternalSetupPortal)
	cmdInternal.AddCommand(cmdInternalBreakingChange)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateK8SToDB)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateResources)

	cmdInternalSetupPortal.Flags().StringVar(&DatabaseURL, "database-url", "", "Database URL")
	cmdInternalSetupPortal.Flags().StringVar(&DatabaseSchema, "database-schema", "", "Database schema name")
	cmdInternalSetupPortal.Flags().StringVar(&defaultAuthgearDomain, "default-authgear-domain", "", "App default domain")
	cmdInternalSetupPortal.Flags().StringVar(&customAuthgearDomain, "custom-authgear-domain", "", "App custom domain")

	_ = cmdInternalSetupPortal.MarkFlagRequired("default-authgear-domain")
	_ = cmdInternalSetupPortal.MarkFlagRequired("custom-authgear-domain")

	cmdInternalBreakingChangeMigrateK8SToDB.Flags().StringVar(&DatabaseURL, "database-url", "", "Database URL")
	cmdInternalBreakingChangeMigrateK8SToDB.Flags().StringVar(&DatabaseSchema, "database-schema", "", "Database schema name")
	cmdInternalBreakingChangeMigrateK8SToDB.Flags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path to kubeconfig")
	cmdInternalBreakingChangeMigrateK8SToDB.Flags().StringVar(&namespace, "namespace", "", "Namespace")

	// FIXME: Respect KUBECONFIG environment variable.
	_ = cmdInternalBreakingChangeMigrateK8SToDB.MarkFlagRequired("kubeconfig")
	_ = cmdInternalBreakingChangeMigrateK8SToDB.MarkFlagRequired("namespace")
}
