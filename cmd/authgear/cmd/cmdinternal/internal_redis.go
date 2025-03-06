package cmdinternal

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	authgearcmd "github.com/authgear/authgear-server/cmd/authgear/cmd"
	cmdredis "github.com/authgear/authgear-server/cmd/authgear/redis"
)

func init() {
	binder := authgearcmd.GetBinder()

	cmdInternal.AddCommand(cmdInternalRedis)
	cmdInternalRedis.AddCommand(cmdInternalRedisCleanUpNonExpiringKeys)

	binder.BindString(cmdInternalRedisCleanUpNonExpiringKeys.Flags(), authgearcmd.ArgRedisURL)
	_ = cmdInternalRedisCleanUpNonExpiringKeys.Flags().Bool("dry-run", true, "Dry-run or not.")
	_ = cmdInternalRedisCleanUpNonExpiringKeys.Flags().String("scan-count", "100", "Redis SCAN count")
	_ = cmdInternalRedisCleanUpNonExpiringKeys.Flags().String("expiration", "", "A Go duration literal.")
	_ = cmdInternalRedisCleanUpNonExpiringKeys.Flags().String("key-pattern", "", fmt.Sprintf("One of %v", strings.Join(cmdredis.KnownKeyPatterns, ",")))
}

var cmdInternalRedis = &cobra.Command{
	Use:   "redis",
	Short: "Redis commands",
}

// See https://linear.app/authgear/issue/DEV-1325/properly-clean-up-redis-keys
var cmdInternalRedisCleanUpNonExpiringKeys = &cobra.Command{
	Use:   "clean-up-non-expiring-keys",
	Short: "Clean up non-expiring keys.",
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
		expirationString := cmd.Flags().Lookup("expiration").Value.String()
		keyPattern := cmd.Flags().Lookup("key-pattern").Value.String()
		dryRun := cmd.Flags().Lookup("dry-run").Value.String() == "true"
		now := time.Now().UTC()

		err = cmdredis.CleanUpNonExpiringKeys(cmd.Context(), redisClient, os.Stdout, log.Default(), cmdredis.CleanUpNonExpiringKeysOptions{
			ScanCountString:  scanCountString,
			KeyPattern:       keyPattern,
			ExpirationString: expirationString,
			DryRun:           dryRun,
			Now:              now,
		})
		if err != nil {
			return err
		}

		return nil
	},
}
