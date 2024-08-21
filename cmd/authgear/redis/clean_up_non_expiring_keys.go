package redis

import (
	"context"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

	goredis "github.com/go-redis/redis/v8"
)

func accessEventsToSession(key string) string {
	return strings.ReplaceAll(key, ":access-events:", ":session:")
}

func accessEventsToOfflineGrant(key string) string {
	return strings.ReplaceAll(key, ":access-events:", ":offline-grant:")
}

func CleanUpNonExpiringKeys(ctx context.Context, redisClient *goredis.Client, dryRun bool, stdout io.Writer, logger *log.Logger) (err error) {
	conn := redisClient.Conn(ctx)
	defer conn.Close()

	// We first scan the key pattern "app:*:access-events:*"
	// For each key, we see if it has a corresponding "app:*:session:*" or "app:*:offline-grant:*"
	// If it does, then the session does not expire and its access-events should be kept.
	// Otherwise, the session has expired, and its access-events can be deleted.

	// We start with a cursor of 0.
	var cursor uint64
	pattern := "app:*:access-events:*"
	// We scan 100 keys
	var count int64 = 100

	var keysToBeDelete []string
	var scannedKeyCount = 0
	for {
		var keys []string
		var nextCursor uint64
		keys, nextCursor, err = conn.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return
		}

		scannedKeyCount += len(keys)
		logger.Printf("SCAN with cursor %v: %v\n", cursor, scannedKeyCount)

		for _, key := range keys {
			offlineGrantKey := accessEventsToOfflineGrant(key)
			sessionKey := accessEventsToSession(key)

			var existsCount int64
			existsCount, err = conn.Exists(ctx, offlineGrantKey, sessionKey).Result()
			if err != nil {
				return
			}

			// No keys exist. The session is gone. This key can be deleted.
			if existsCount == 0 {
				keysToBeDelete = append(keysToBeDelete, key)
			}
		}

		// According to the doc https://redis.io/docs/latest/commands/scan/
		// a cursor of 0 means the scan is complete.
		if nextCursor == 0 {
			logger.Printf("done SCAN: %v\n", scannedKeyCount)
			break
		} else {
			cursor = nextCursor
		}
	}

	// Sort the keys.
	slices.Sort(keysToBeDelete)

	for _, key := range keysToBeDelete {
		if !dryRun {
			_, err = conn.Del(ctx, key).Result()
			if err != nil {
				return
			}
			logger.Printf("deleted %v\n", key)
		} else {
			logger.Printf("would delete %v\n", key)
		}
		// We prints to stdout anyway.
		fmt.Fprintf(stdout, "%v\n", key)
	}

	return nil
}
