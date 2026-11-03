package cmdinternal

import (
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdredis "github.com/authgear/authgear-server/cmd/authgear/redis"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalRedis)

	cmdInternalRedis.AddCommand(cmdInternalRedisHousekeep)
	binder.BindString(cmdInternalRedisHousekeep.Flags(), authgearcmd.ArgRedisURL)
	_ = cmdInternalRedisHousekeep.Flags().Bool("dry-run", true, "Dry-run or not.")
	_ = cmdInternalRedisHousekeep.Flags().String("scan-count", "10000", "Redis SCAN count, which is roughly the number of keys handled per batch")
}

var cmdInternalRedis = &cobra.Command{
	Use:   "redis",
	Short: "Redis commands",
}

// See https://linear.app/authgear/issue/DEV-3995
var cmdInternalRedisHousekeep = &cobra.Command{
	Use:   "housekeep",
	Short: "Delete or fix the expiry of Redis keys that are known to leak. Safe to run again.",
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

		scanCountString := cmd.Flags().Lookup("scan-count").Value.String()
		dryRun := cmd.Flags().Lookup("dry-run").Value.String() == "true"
		now := time.Now().UTC()

		return cmdredis.Housekeep(cmd.Context(), redisClient, os.Stdout, log.Default(), cmdredis.HousekeepOptions{
			ScanCountString: scanCountString,
			DryRun:          dryRun,
			Now:             now,
		})
	},
}
