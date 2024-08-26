package redis

import (
	"context"
	"fmt"
	"io"
	"log"
	"slices"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

func ListNonExpiringKeys(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger) (err error) {
	conn := redisClient.Conn(ctx)
	defer conn.Close()

	// We start with a cursor of 0.
	var cursor uint64
	// pattern is "*" so that we SCAN all keys.
	pattern := "*"
	// We scan 100 keys
	var count int64 = 100

	var nonExpiringKeys []string
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

		// And then we call TTL on the key.
		for _, key := range keys {
			var duration time.Duration
			duration, err = conn.TTL(ctx, key).Result()
			if err != nil {
				return
			}
			switch duration {
			case -2:
				// -2 means the key does not exist.
				// Since it no longer exists, I assume it expired already.
				break
			case -1:
				// -1 means the key does not expire.
				nonExpiringKeys = append(nonExpiringKeys, key)
			default:
				// Other values mean the key does have TTL and is going to expire some time in the future.
				break
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

	// Sort the keys before printing.
	slices.Sort(nonExpiringKeys)
	for _, key := range nonExpiringKeys {
		fmt.Fprintf(stdout, "%v\n", key)
	}

	return nil
}
