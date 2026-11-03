package redis

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Matches both the global and the per-app key formats in pkg/lib/ratelimit/limiter.go.
const rateLimitScanPattern = "*rate-limit:*"

var rateLimitKeyRegexp = regexp.MustCompile(`^(rate-limit:|app:[^:]+:rate-limit:)`)

// A rate-limit key stores the time in milliseconds when its bucket is full again.
// Before DEV-3995, the key was given that time to EXPIREAT as seconds, so it never expired.
// A key past that time holds no state and is deleted; otherwise it is given the intended expiry.
// The check and the write run in one script, so a key updated in between is not deleted.
var cleanUpRateLimitKeysScript = goredis.NewScript(`
local now = tonumber(ARGV[1])
local dry_run = ARGV[2] == "1"
local results = {}
for i, key in ipairs(KEYS) do
	local result = "skipped"
	local value = redis.pcall("GET", key)
	local t = nil
	if type(value) == "string" then
		t = tonumber(value)
	end
	if t then
		if t < now then
			if not dry_run then
				redis.call("DEL", key)
			end
			result = "deleted"
		else
			if not dry_run then
				redis.call("PEXPIREAT", key, math.floor(t))
			end
			result = "expiry-fixed"
		end
	end
	results[i] = result
end
return results
`)

type CleanUpRateLimitKeysOptions struct {
	ScanCountString string
	DryRun          bool
	Now             time.Time
}

func CleanUpRateLimitKeys(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger, options CleanUpRateLimitKeysOptions) (err error) {
	scanCount, err := strconv.ParseInt(options.ScanCountString, 10, 64)
	if err != nil {
		err = fmt.Errorf("SCAN count must be an integer: %v", options.ScanCountString)
		return
	}
	if scanCount < 10 {
		err = fmt.Errorf("SCAN count must greater than or equal to 10: %v", scanCount)
		return
	}

	if options.Now.IsZero() {
		err = fmt.Errorf("now cannot be zero")
		return
	}

	dryRun := "0"
	dryRunPrefix := ""
	if options.DryRun {
		dryRun = "1"
		dryRunPrefix = "(dry-run) "
	}
	now := strconv.FormatInt(options.Now.UnixMilli(), 10)

	conn := redisClient.Conn()
	defer func() {
		_ = conn.Close()
	}()

	var cursor uint64
	counts := map[string]int{}
	scannedKeyCount := 0
	for {
		var keys []string
		var nextCursor uint64
		keys, nextCursor, err = conn.Scan(ctx, cursor, rateLimitScanPattern, scanCount).Result()
		if err != nil {
			return
		}

		var rateLimitKeys []string
		for _, key := range keys {
			if rateLimitKeyRegexp.MatchString(key) {
				rateLimitKeys = append(rateLimitKeys, key)
			}
		}
		scannedKeyCount += len(rateLimitKeys)

		if len(rateLimitKeys) > 0 {
			var results []string
			results, err = cleanUpRateLimitKeysScript.Run(ctx, conn, rateLimitKeys, now, dryRun).StringSlice()
			if err != nil {
				return
			}
			for i, result := range results {
				counts[result] += 1
				if result != "skipped" {
					fmt.Fprintf(stdout, "%v%v %v\n", dryRunPrefix, result, rateLimitKeys[i])
				}
			}
		}

		logger.Printf("%vSCAN %v %v scanned_total=%v deleted_total=%v expiry_fixed_total=%v skipped_total=%v\n",
			dryRunPrefix, cursor, rateLimitScanPattern, scannedKeyCount, counts["deleted"], counts["expiry-fixed"], counts["skipped"])

		if nextCursor == 0 {
			logger.Printf("%vdone scanned_total=%v deleted_total=%v expiry_fixed_total=%v skipped_total=%v\n",
				dryRunPrefix, scannedKeyCount, counts["deleted"], counts["expiry-fixed"], counts["skipped"])
			break
		}
		cursor = nextCursor
	}

	return nil
}
