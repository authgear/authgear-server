package cmdinternal

import (
	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalRedis)
	cmdInternalRedis.AddCommand(cmdInternalRedisListNonExpiringKeys)
	cmdInternalRedis.AddCommand(cmdInternalRedisCleanUpNonExpiringKeys)

	binder.BindString(cmdInternalRedisListNonExpiringKeys.Flags(), authgearcmd.ArgRedisURL)

	binder.BindString(cmdInternalRedisCleanUpNonExpiringKeys.Flags(), authgearcmd.ArgRedisURL)
	_ = cmdInternalRedisCleanUpNonExpiringKeys.Flags().Bool("dry-run", true, "Dry-run or not.")
}

var cmdInternalRedis = &cobra.Command{
	Use:   "redis",
	Short: "Redis commands",
}

var cmdInternalRedisListNonExpiringKeys = &cobra.Command{
	Use:   "list-non-expiring-keys",
	Short: "List all non-expiring keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		var err error
		_, err = binder.GetRequiredString(cmd, authgearcmd.ArgRedisURL)
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalRedisCleanUpNonExpiringKeys = &cobra.Command{
	Use:   "clean-up-non-expiring-keys",
	Short: "Clean up all known non-expiring keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		var err error
		_, err = binder.GetRequiredString(cmd, authgearcmd.ArgRedisURL)
		if err != nil {
			return err
		}

		return nil
	},
}
