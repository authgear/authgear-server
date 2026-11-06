package redis

import (
	"context"
	"regexp"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// The key formats in pkg/lib/lockout/storage_redis.go and pkg/lib/authn/otp/attempt_tracker_redis.go.
var lockoutKeyRegexp = regexp.MustCompile(`^app:[^:]+:lockout:`)
var failedAttemptsKeyRegexp = regexp.MustCompile(`^app:[^:]+:failed-attempts:`)

// The application always gives these keys an expiry,
// so a key without one is left over and holds a stale counter.
var housekeepTaskLockout = housekeepTask{
	name:  "lockout",
	match: lockoutKeyRegexp.MatchString,
	plan:  planDeleteKeysWithoutExpiry,
}

var housekeepTaskFailedAttempts = housekeepTask{
	name:  "failed-attempts",
	match: failedAttemptsKeyRegexp.MatchString,
	plan:  planDeleteKeysWithoutExpiry,
}

func planDeleteKeysWithoutExpiry(ctx context.Context, r housekeepReader, keys []string, now time.Time) ([]housekeepAction, error) {
	ttls := make([]*goredis.DurationCmd, len(keys))
	err := readPipelined(ctx, r, func(pipe goredis.Pipeliner) {
		for i, key := range keys {
			ttls[i] = pipe.TTL(ctx, key)
		}
	})
	if err != nil {
		return nil, err
	}

	var actions []housekeepAction
	for i, key := range keys {
		ttl, err := ttls[i].Result()
		if err != nil {
			return nil, err
		}
		// -1 means the key has no expiry.
		if ttl == -1 {
			actions = append(actions, deleteAction(key))
		}
	}
	return actions, nil
}
