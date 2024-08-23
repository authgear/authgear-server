package cmdinternal

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdredis "github.com/authgear/authgear-server/cmd/authgear/redis"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalRedis)
	cmdInternalRedis.AddCommand(cmdInternalRedisListNonExpiringKeys)
	cmdInternalRedis.AddCommand(cmdInternalRedisCleanUpNonExpiringKeys)

	cmdInternalRedisCleanUpNonExpiringKeys.AddCommand(cmdInternalRedisCleanUpNonExpiringKeysAccessEvents)
	cmdInternalRedisCleanUpNonExpiringKeys.AddCommand(cmdInternalRedisCleanUpNonExpiringKeysSessionHashes)

	binder.BindString(cmdInternalRedisListNonExpiringKeys.Flags(), authgearcmd.ArgRedisURL)

	binder.BindString(cmdInternalRedisCleanUpNonExpiringKeysAccessEvents.Flags(), authgearcmd.ArgRedisURL)
	_ = cmdInternalRedisCleanUpNonExpiringKeysAccessEvents.Flags().Bool("dry-run", true, "Dry-run or not.")

	binder.BindString(cmdInternalRedisCleanUpNonExpiringKeysSessionHashes.Flags(), authgearcmd.ArgRedisURL)
	_ = cmdInternalRedisCleanUpNonExpiringKeysSessionHashes.Flags().Bool("dry-run", true, "Dry-run or not.")
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

		redisURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgRedisURL)
		if err != nil {
			return err
		}

		redisClient, err := cmdredis.NewClient(redisURL)
		if err != nil {
			return err
		}

		ctx := context.Background()
		err = cmdredis.ListNonExpiringKeys(ctx, redisClient, os.Stdout, log.Default())
		if err != nil {
			return err
		}

		return nil
	},
}

// According to https://linear.app/authgear/issue/DEV-1325/properly-clean-up-redis-keys
// There are some known key patterns that are non-expiring.
//
//	failed-attempts
//	session-list
//	access-events
//	lockout
//	offline-grant-list
//
// For failed-attempts, lockout, the number of them should be small enough to be ignored.
// They can also be deleted with redis-cli, given we have list-non-expiring-keys.
//
//	$ authgear internal redis list-non-expiring-keys >output
//	$ grep <output 'failed-attempts|lockout' >targets
//	$ xargs <targets redis-cli del
//
// For access-events, it is handled by the subcommand access-events.
// For session-list and offline-grant-list, it is handled by the subcommand session-hashes.
var cmdInternalRedisCleanUpNonExpiringKeys = &cobra.Command{
	Use:   "clean-up-non-expiring-keys",
	Short: "Clean up non-expiring keys.",
}

var cmdInternalRedisCleanUpNonExpiringKeysAccessEvents = &cobra.Command{
	Use:   "access-events",
	Short: "Clean up non-expiring 'app:*:access-events:*'",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		redisURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgRedisURL)
		if err != nil {
			return err
		}

		redisClient, err := cmdredis.NewClient(redisURL)
		if err != nil {
			return err
		}

		ctx := context.Background()
		dryRun := cmd.Flags().Lookup("dry-run").Value.String() == "true"
		err = cmdredis.CleanUpNonExpiringKeysAccessEvents(ctx, redisClient, dryRun, os.Stdout, log.Default())
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdInternalRedisCleanUpNonExpiringKeysSessionHashes = &cobra.Command{
	Use:   "session-hashes",
	Short: "Clean up non-expiring 'app:*:session-list:*' and 'app:*:offline-grant-list:*'",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := authgearcmd.GetBinder()

		redisURL, err := binder.GetRequiredString(cmd, authgearcmd.ArgRedisURL)
		if err != nil {
			return err
		}

		redisClient, err := cmdredis.NewClient(redisURL)
		if err != nil {
			return err
		}

		ctx := context.Background()
		dryRun := cmd.Flags().Lookup("dry-run").Value.String() == "true"
		err = cmdredis.CleanUpNonExpiringKeysSessionHashes(ctx, redisClient, dryRun, os.Stdout, log.Default())
		if err != nil {
			return err
		}

		return nil
	},
}
