package cmdinternal

import (
	"log"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			log.Fatalf(err.Error())
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			log.Fatalf(err.Error())
		}
		defaultAuthgearDomain, err := binder.GetRequiredString(cmd, portalcmd.ArgDefaultAuthgearDomain)
		if err != nil {
			log.Fatalf(err.Error())
		}
		customAuthgearDomain, err := binder.GetRequiredString(cmd, portalcmd.ArgCustomAuthgearDomain)
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
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		kubeConfigPath := binder.GetString(cmd, portalcmd.ArgKubeconfig)

		namespace, err := binder.GetRequiredString(cmd, portalcmd.ArgNamespace)
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
	binder := portalcmd.GetBinder()
	cmdInternal.AddCommand(cmdInternalSetupPortal)
	cmdInternal.AddCommand(cmdInternalBreakingChange)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateK8SToDB)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateResources)

	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgDefaultAuthgearDomain)
	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgCustomAuthgearDomain)

	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgKubeconfig)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgNamespace)

	portalcmd.Root.AddCommand(cmdInternal)
}
