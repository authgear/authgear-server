package redis

import (
	"context"
	"fmt"
	"io"
	"log"
	"slices"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var KnownKeyPatterns = []string{
	"session-list",
	"offline-grant-list",
	"failed-attempts",
	"access-events",
	"lockout",
}

type CleanUpNonExpiringKeysOptions struct {
	ScanCountString  string
	KeyPattern       string
	ExpirationString string
	DryRun           bool
	Now              time.Time
}

func toRedisKeyPattern(keyPattern string) string {
	return fmt.Sprintf("app:*:%v:*", keyPattern)
}

func isCandidateForExpire(ctx context.Context, conn *goredis.Conn, key string, isSessionHash bool, now time.Time) (ok bool, err error) {
	if !isSessionHash {
		var ttl time.Duration
		ttl, err = conn.TTL(ctx, key).Result()
		if err != nil {
			return
		}
		switch ttl {
		case -2:
			// -2 means the key does not exist.
			// Just ignore it.
			break
		case -1:
			// -1 means the key does not expire.
			ok = true
		default:
			// Other values mean the key does have TTL and is going to expire some time in the future.
			break
		}
	} else {
		var hash map[string]string
		hash, err = conn.HGetAll(ctx, key).Result()
		if err != nil {
			return
		}

		// The hash has no entries.
		// So it can expire.
		if len(hash) == 0 {
			ok = true
		} else {
			hasSessionThatWillExpireInTheFuture := false
			// The values are time.Time.MarshalText()
			// We check if all values are timestamp in the past.
			// If yes, then it can expire.
			for _, rfc3339 := range hash {
				var t time.Time
				err = t.UnmarshalText([]byte(rfc3339))
				if err != nil {
					return
				}
				if t.After(now) {
					hasSessionThatWillExpireInTheFuture = true
				}
			}
			if !hasSessionThatWillExpireInTheFuture {
				ok = true
			}
		}
	}
	return
}

func CleanUpNonExpiringKeys(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger, options CleanUpNonExpiringKeysOptions) (err error) {
	// https://redis.io/docs/latest/commands/scan/
	// The default is 10.
	// It makes no sense to use a value less than 10.
	scanCount, err := strconv.ParseInt(options.ScanCountString, 10, 64)
	if err != nil {
		err = fmt.Errorf("SCAN count must be an integer: %v", options.ScanCountString)
		return
	}
	if scanCount < 10 {
		err = fmt.Errorf("SCAN count must greater than or equal to 10: %v", scanCount)
		return
	}

	expiration, err := time.ParseDuration(options.ExpirationString)
	if err != nil {
		err = fmt.Errorf("expiration must be a valid Go duration literal: %v", options.ExpirationString)
		return
	}
	if expiration < 0 {
		err = fmt.Errorf("expiration cannot be less than 0: %v", expiration)
		return
	}

	if !slices.Contains(KnownKeyPatterns, options.KeyPattern) {
		err = fmt.Errorf("unsupported key patttern: %v", options.KeyPattern)
		return
	}

	if options.Now.IsZero() {
		err = fmt.Errorf("now cannot be zero")
		return
	}

	isSessionHash := false
	switch options.KeyPattern {
	case "session-list":
		isSessionHash = true
	case "offline-grant-list":
		isSessionHash = true
	}

	conn := redisClient.Conn()
	defer func() {
		_ = conn.Close()
	}()

	// We start with a cursor of 0.
	var cursor uint64
	pattern := toRedisKeyPattern(options.KeyPattern)

	var scannedKeyCount = 0
	var expiredKeyCount = 0
	for {
		var keys []string
		var nextCursor uint64
		keys, nextCursor, err = conn.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			return
		}

		scannedKeyCount += len(keys)
		logger.Printf("SCAN %v %v scanned_total=%v expired_total=%v\n", cursor, pattern, scannedKeyCount, expiredKeyCount)

		for _, key := range keys {
			var ok bool
			ok, err = isCandidateForExpire(ctx, conn, key, isSessionHash, options.Now)
			if err != nil {
				return
			}
			if ok {
				expiredKeyCount += 1
				// -1 means the key does not expire.
				if options.DryRun {
					logger.Printf("(dry-run) EXPIRE %v %v\n", key, expiration.Seconds())
				} else {
					_, err = conn.Expire(ctx, key, expiration).Result()
					if err != nil {
						return
					}
					logger.Printf("EXPIRE %v %v\n", key, expiration.Seconds())
				}
				fmt.Fprintf(stdout, "%v\n", key)
			}
		}

		// According to the doc https://redis.io/docs/latest/commands/scan/
		// a cursor of 0 means the scan is complete.
		if nextCursor == 0 {
			logger.Printf("done scanned_total=%v expired_total=%v\n", scannedKeyCount, expiredKeyCount)
			break
		} else {
			cursor = nextCursor
		}
	}

	return nil
}
