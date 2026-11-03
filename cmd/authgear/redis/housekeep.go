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

type housekeepKind struct {
	name      string
	keyRegexp *regexp.Regexp
}

// The key formats are defined in
// pkg/lib/ratelimit/limiter.go, pkg/lib/session/access (removed in DEV-3995),
// pkg/lib/lockout/storage_redis.go and pkg/lib/authn/otp/attempt_tracker_redis.go.
var housekeepKinds = []housekeepKind{
	{name: "rate-limit", keyRegexp: regexp.MustCompile(`^(rate-limit:|app:[^:]+:rate-limit:)`)},
	{name: "access-events", keyRegexp: regexp.MustCompile(`^app:[^:]+:access-events:`)},
	{name: "lockout", keyRegexp: regexp.MustCompile(`^app:[^:]+:lockout:`)},
	{name: "failed-attempts", keyRegexp: regexp.MustCompile(`^app:[^:]+:failed-attempts:`)},
}

const (
	housekeepResultDeleted     = "deleted"
	housekeepResultExpiryFixed = "expiry-fixed"
	housekeepResultSkipped     = "skipped"
)

// KEYS are the keys of one batch, ARGV[1] is now in milliseconds, ARGV[2] is "1" for dry-run,
// and ARGV[2+i] is the kind of KEYS[i]. Each key is checked and changed in the same script,
// so a key written by the application in between is never changed by mistake.
//
// rate-limit: The key stores the time in milliseconds when its bucket is full again.
// Before DEV-3995, it was given to EXPIREAT as seconds, so the key never expired.
// A key past that time holds no state and is deleted; otherwise it is given the intended expiry.
//
// access-events: The stream was written but never read, and is no longer written since DEV-3995.
//
// lockout, failed-attempts: The application always gives them an expiry,
// so a key without one is left over and holds a stale counter.
var housekeepScript = goredis.NewScript(`
local now = tonumber(ARGV[1])
local dry_run = ARGV[2] == "1"
local results = {}

local function del(key)
	if not dry_run then
		redis.call("DEL", key)
	end
	return "deleted"
end

for i, key in ipairs(KEYS) do
	local kind = ARGV[i + 2]
	local result = "skipped"

	if kind == "rate-limit" then
		local value = redis.pcall("GET", key)
		local t = nil
		if type(value) == "string" then
			t = tonumber(value)
		end
		if t then
			if t < now then
				result = del(key)
			else
				-- A correct expiry is left alone, so a rerun reports nothing.
				-- The tolerance absorbs clock skew between this machine and Redis.
				local pttl = redis.call("PTTL", key)
				if pttl == -1 or now + pttl > t + 60000 then
					if not dry_run then
						redis.call("PEXPIREAT", key, math.floor(t))
					end
					result = "expiry-fixed"
				end
			end
		end
	elseif kind == "access-events" then
		local key_type = redis.call("TYPE", key)
		if type(key_type) == "table" then
			key_type = key_type.ok
		end
		if key_type == "stream" then
			result = del(key)
		end
	elseif kind == "lockout" or kind == "failed-attempts" then
		if redis.call("TTL", key) == -1 then
			result = del(key)
		end
	end

	results[i] = result
end
return results
`)

type HousekeepOptions struct {
	ScanCountString string
	DryRun          bool
	Now             time.Time
}

type housekeepCounts struct {
	deleted     int
	expiryFixed int
	skipped     int
}

func (c *housekeepCounts) add(result string) {
	switch result {
	case housekeepResultDeleted:
		c.deleted += 1
	case housekeepResultExpiryFixed:
		c.expiryFixed += 1
	default:
		c.skipped += 1
	}
}

func (c housekeepCounts) String() string {
	return fmt.Sprintf("deleted=%v expiry_fixed=%v skipped=%v", c.deleted, c.expiryFixed, c.skipped)
}

func classifyHousekeepKey(key string) (kind string, ok bool) {
	for _, k := range housekeepKinds {
		if k.keyRegexp.MatchString(key) {
			return k.name, true
		}
	}
	return "", false
}

// Housekeep scans the whole keyspace once, one SCAN batch at a time, so memory use does not grow with the keyspace.
// It only changes keys that are known to be safe to change, so it can be run again at any time.
func Housekeep(ctx context.Context, redisClient *goredis.Client, stdout io.Writer, logger *log.Logger, options HousekeepOptions) (err error) {
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
	scannedKeyCount := 0
	total := housekeepCounts{}
	countsByKind := map[string]*housekeepCounts{}
	for _, k := range housekeepKinds {
		countsByKind[k.name] = &housekeepCounts{}
	}

	for {
		var keys []string
		var nextCursor uint64
		keys, nextCursor, err = conn.Scan(ctx, cursor, "", scanCount).Result()
		if err != nil {
			return
		}
		scannedKeyCount += len(keys)

		var targetKeys []string
		args := []any{now, dryRun}
		for _, key := range keys {
			if kind, ok := classifyHousekeepKey(key); ok {
				targetKeys = append(targetKeys, key)
				args = append(args, kind)
			}
		}

		if len(targetKeys) > 0 {
			var results []string
			results, err = housekeepScript.Run(ctx, conn, targetKeys, args...).StringSlice()
			if err != nil {
				return
			}
			for i, result := range results {
				kind := args[i+2].(string)
				countsByKind[kind].add(result)
				total.add(result)
				if result != housekeepResultSkipped {
					fmt.Fprintf(stdout, "%v%v %v\n", dryRunPrefix, result, targetKeys[i])
				}
			}
		}

		logger.Printf("%vSCAN %v scanned_total=%v %v\n", dryRunPrefix, cursor, scannedKeyCount, total)

		if nextCursor == 0 {
			for _, k := range housekeepKinds {
				logger.Printf("%vdone %v %v\n", dryRunPrefix, k.name, countsByKind[k.name])
			}
			logger.Printf("%vdone scanned_total=%v %v\n", dryRunPrefix, scannedKeyCount, total)
			break
		}
		cursor = nextCursor
	}

	return nil
}
