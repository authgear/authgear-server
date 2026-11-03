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
	plan: func(ctx context.Context, tx *goredis.Tx, keys []string, now time.Time) ([]housekeepAction, error) {
		gets := make([]*goredis.StringCmd, len(keys))
		pttls := make([]*goredis.DurationCmd, len(keys))
		err := readPipelined(ctx, tx, func(pipe goredis.Pipeliner) {
			for i, key := range keys {
				gets[i] = pipe.Get(ctx, key)
				pttls[i] = pipe.PTTL(ctx, key)
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
			ms, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			fullAt := time.UnixMilli(int64(ms)).UTC()

			if fullAt.Before(now) {
				actions = append(actions, deleteAction(key))
				continue
			}

			pttl, err := pttls[i].Result()
			if err != nil {
				return nil, err
			}
			// -1 means the key has no expiry.
			if pttl == -1 || now.Add(pttl).After(fullAt.Add(rateLimitExpiryTolerance)) {
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
