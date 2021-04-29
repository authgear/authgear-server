package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternal = &cobra.Command{
	Use:   "internal [setup-portal]",
	Short: "Setup portal config source data in db",
}

var cmdInternalSetupPortal = &cobra.Command{
	Use:   "setup-portal",
	Short: "Initialize app configuration",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		defaultAuthgearDomain, err := ArgDefaultAuthgearDomain.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
		}
		customAuthgearDomain, err := ArgCustomAuthgearDomain.GetRequired(viper.GetViper())
		if err != nil {
			log.Fatalf(err.Error())
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
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			return err
		}

		kubeConfigPath := ArgKubeconfig.Get(viper.GetViper())

		namespace, err := ArgNamespace.GetRequired(viper.GetViper())
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

	ArgDatabaseURL.Bind(cmdInternalSetupPortal.Flags(), viper.GetViper())
	ArgDatabaseSchema.Bind(cmdInternalSetupPortal.Flags(), viper.GetViper())
	ArgDefaultAuthgearDomain.Bind(cmdInternalSetupPortal.Flags(), viper.GetViper())
	ArgCustomAuthgearDomain.Bind(cmdInternalSetupPortal.Flags(), viper.GetViper())

	ArgDatabaseURL.Bind(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), viper.GetViper())
	ArgDatabaseSchema.Bind(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), viper.GetViper())
	ArgKubeconfig.Bind(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), viper.GetViper())
	ArgNamespace.Bind(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), viper.GetViper())
}
