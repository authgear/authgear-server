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
// For session-list and offline-grant-list, they will be deleted when the user is deleted.
//
// So access-events is the only key pattern left. The amount of them is huge.
// And the deletion logic is not trivial so this command handles this case.
var cmdInternalRedisCleanUpNonExpiringKeys = &cobra.Command{
	Use:   "clean-up-non-expiring-keys",
	Short: "Clean up all known non-expiring keys. Currently it only handles access-events",
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
		err = cmdredis.CleanUpNonExpiringKeys(ctx, redisClient, dryRun, os.Stdout, log.Default())
		if err != nil {
			return err
		}

		return nil
	},
}
