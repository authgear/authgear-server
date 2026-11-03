package redis

import (
	"context"
	"regexp"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Both the global and the per-app key formats in pkg/lib/ratelimit/limiter.go.
var rateLimitKeyRegexp = regexp.MustCompile(`^(rate-limit:|app:[^:]+:rate-limit:)`)

// The tolerance absorbs clock skew between this machine and Redis.
const rateLimitExpiryTolerance = time.Minute

// A rate limit key stores the time in milliseconds when its bucket is full again.
// Before DEV-3995, it was given to EXPIREAT as seconds, so the key never expired.
// A key past that time holds no state and is deleted; otherwise it is given the intended expiry.
// A key that already has the intended expiry is left alone.
var housekeepTaskRateLimit = housekeepTask{
	name:  "rate-limit",
	match: rateLimitKeyRegexp.MatchString,
	plan: func(ctx context.Context, r housekeepReader, keys []string, now time.Time) ([]housekeepAction, error) {
		gets := make([]*goredis.StringCmd, len(keys))
		// PTTL is read as a plain integer, because the wrong expiry is tens of thousands of years away,
		// which overflows time.Duration.
		pttls := make([]*goredis.Cmd, len(keys))
		err := readPipelined(ctx, r, func(pipe goredis.Pipeliner) {
			for i, key := range keys {
				gets[i] = pipe.Get(ctx, key)
				pttls[i] = pipe.Do(ctx, "PTTL", key)
			}
		})
		if err != nil {
			return nil, err
		}

		var actions []housekeepAction
		for i, key := range keys {
			value, err := gets[i].Result()
			if skip, err := housekeepCmdErr(err); err != nil {
				return nil, err
			} else if skip {
				continue
			}
			pttlMs, err := pttls[i].Int64()
			if err != nil {
				return nil, err
			}

			result, fullAt := planRateLimitKey(value, pttlMs, now)
			switch result {
			case housekeepResultDeleted:
				actions = append(actions, deleteAction(key))
			case housekeepResultExpiryFixed:
				actions = append(actions, housekeepAction{
					key:    key,
					result: housekeepResultExpiryFixed,
					apply: func(ctx context.Context, pipe goredis.Pipeliner) {
						pipe.PExpireAt(ctx, key, fullAt)
					},
				})
			}
		}
		return actions, nil
	},
}

// planRateLimitKey decides what to do with a rate limit key from its value and its PTTL in milliseconds.
// It returns an empty result if the key is left alone.
func planRateLimitKey(value string, pttlMs int64, now time.Time) (result string, fullAt time.Time) {
	ms, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", time.Time{}
	}
	fullAt = time.UnixMilli(int64(ms)).UTC()

	if fullAt.Before(now) {
		return housekeepResultDeleted, fullAt
	}

	// -1 means the key has no expiry.
	expireAtMs := now.UnixMilli() + pttlMs
	if pttlMs == -1 || expireAtMs > fullAt.Add(rateLimitExpiryTolerance).UnixMilli() {
		return housekeepResultExpiryFixed, fullAt
	}
	return "", fullAt
}
