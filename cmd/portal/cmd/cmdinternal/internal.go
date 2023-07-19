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

var cmdInternalSetupPortal = &cobra.Command{
	Use:   "setup-portal",
	Short: "Initialize app configuration",
	Args:  cobra.ExactArgs(1),
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

		resourceDir := args[0]

		err = internal.SetupPortal(&internal.SetupPortalOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			ResourceDir:    resourceDir,
		})
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalUnpack = &cobra.Command{
	Use:   "unpack",
	Short: "Unpack database configsource data JSON to a directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()
		dataJSONPath, err := binder.GetRequiredString(cmd, portalcmd.ArgDataJSONFilePath)
		if err != nil {
			return err
		}

		outputDirectoryPath, err := binder.GetRequiredString(cmd, portalcmd.ArgOutputDirectoryPath)
		if err != nil {
			return err
		}

		return internal.Unpack(&internal.UnpackOptions{
			DataJSONPath:        dataJSONPath,
			OutputDirectoryPath: outputDirectoryPath,
		})
	},
}

var cmdInternalPack = &cobra.Command{
	Use:   "pack",
	Short: "Pack unpacked directory into database configsource data JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()

		inputDirectoryPath, err := binder.GetRequiredString(cmd, portalcmd.ArgInputDirectoryPath)
		if err != nil {
			return err
		}

		return internal.Pack(&internal.PackOptions{
			InputDirectoryPath: inputDirectoryPath,
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
	cmdInternal.AddCommand(cmdInternalUnpack)
	cmdInternal.AddCommand(cmdInternalPack)
	cmdInternal.AddCommand(cmdInternalBreakingChange)
	cmdInternal.AddCommand(cmdInternalCheck)

	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateK8SToDB)
	cmdInternalBreakingChange.AddCommand(cmdInternalBreakingChangeMigrateResources)

	cmdInternalCheck.AddCommand(cmdInternalCheckConfigSources)

	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalSetupPortal.Flags(), portalcmd.ArgDatabaseSchema)

	binder.BindString(cmdInternalUnpack.Flags(), portalcmd.ArgDataJSONFilePath)
	binder.BindString(cmdInternalUnpack.Flags(), portalcmd.ArgOutputDirectoryPath)

	binder.BindString(cmdInternalPack.Flags(), portalcmd.ArgInputDirectoryPath)

	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgKubeconfig)
	binder.BindString(cmdInternalBreakingChangeMigrateK8SToDB.Flags(), portalcmd.ArgNamespace)

	binder.BindString(cmdInternalCheckConfigSources.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdInternalCheckConfigSources.Flags(), portalcmd.ArgDatabaseSchema)

	portalcmd.Root.AddCommand(cmdInternal)
}
