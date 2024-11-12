package cmdinternal

import (
	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalConfigSource = &cobra.Command{
	Use:   "configsource",
	Short: "Config source commands.",
}

var cmdInternalConfigSourceCreate = &cobra.Command{
	Use:   "create",
	Short: "create a config source record in the database with the given config source directory.",
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

		err = internal.Create(cmd.Context(), &internal.CreateOptions{
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

var cmdInternalConfigSourceUnpack = &cobra.Command{
	Use:   "unpack",
	Short: "Unpack a database config source data JSON file into a config source directory.",
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

var cmdInternalConfigSourcePack = &cobra.Command{
	Use:   "pack",
	Short: "Pack a config source directory into a database config source data JSON file.",
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

var cmdInternalConfigSourceCheckDatabase = &cobra.Command{
	Use:   "check-database [app-id ...]",
	Short: "Check the integrity of the config source of the given app IDs. Check all apps if no arguments are given",
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

		err = internal.CheckConfigSources(cmd.Context(), &internal.CheckConfigSourcesOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			AppIDs:         args,
		})

		return err
	},
}
