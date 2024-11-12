package cmdinternal

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternal = &cobra.Command{
	Use:   "internal",
	Short: "Internal commands. Subject to changes without deprecation or removal warnings",
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

		err = internal.MigrateK8SToDB(cmd.Context(), &internal.MigrateK8SToDBOptions{
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

	cmdInternal.AddCommand(cmdInternalConfigSource)
	cmdInternal.AddCommand(cmdInternalDomain)
	cmdInternal.AddCommand(cmdInternalBreakingChange)

	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateK8SToDB)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateResources)

	cmdInternalConfigSource.AddCommand(cmdInternalConfigSourceCreate)
	cmdInternalConfigSource.AddCommand(cmdInternalConfigSourceUnpack)
	cmdInternalConfigSource.AddCommand(cmdInternalConfigSourcePack)
	cmdInternalConfigSource.AddCommand(cmdInternalConfigSourceCheckDatabase)

	cmdInternalDomain.AddCommand(cmdInternalDomainCreateDefault)
	cmdInternalDomain.AddCommand(cmdInternalDomainCreateCustom)

	binder.BindString(cmdInternalConfigSourceCreate.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalConfigSourceCreate.Flags(), portalcmd.ArgDatabaseSchema)

	binder.BindString(cmdInternalConfigSourceUnpack.Flags(), portalcmd.ArgDataJSONFilePath)
	binder.BindString(cmdInternalConfigSourceUnpack.Flags(), portalcmd.ArgOutputDirectoryPath)

	binder.BindString(cmdInternalConfigSourcePack.Flags(), portalcmd.ArgInputDirectoryPath)

	binder.BindString(cmdInternalConfigSourceCheckDatabase.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalConfigSourceCheckDatabase.Flags(), portalcmd.ArgDatabaseSchema)

	binder.BindString(cmdInternalDomainCreateDefault.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalDomainCreateDefault.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalDomainCreateDefault.Flags(), portalcmd.ArgDefaultDomainSuffix)

	binder.BindString(cmdInternalDomainCreateCustom.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalDomainCreateCustom.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalDomainCreateCustom.Flags(), portalcmd.ArgDomain)
	binder.BindString(cmdInternalDomainCreateCustom.Flags(), portalcmd.ArgApexDomain)

	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgKubeconfig)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgNamespace)

	portalcmd.Root.AddCommand(cmdInternal)
}
