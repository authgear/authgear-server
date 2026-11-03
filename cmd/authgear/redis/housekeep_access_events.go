package redis

import (
	"context"
	"regexp"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// The key format of the removed pkg/lib/session/access/event_store_redis.go.
var accessEventsKeyRegexp = regexp.MustCompile(`^app:[^:]+:access-events:`)

// Access event streams were written on every session access but never read.
// They are no longer written since DEV-3995, so every remaining stream is deleted.
var housekeepTaskAccessEvents = housekeepTask{
	name:  "access-events",
	match: accessEventsKeyRegexp.MatchString,
	plan: func(ctx context.Context, r housekeepReader, keys []string, now time.Time) ([]housekeepAction, error) {
		types := make([]*goredis.StatusCmd, len(keys))
		err := readPipelined(ctx, r, func(pipe goredis.Pipeliner) {
			for i, key := range keys {
				types[i] = pipe.Type(ctx, key)
			}
		})
		if err != nil {
			return nil, err
		}

		var actions []housekeepAction
		for i, key := range keys {
			keyType, err := types[i].Result()
			if err != nil {
				return nil, err
			}
			if keyType == "stream" {
				actions = append(actions, deleteAction(key))
			}
		}
		return actions, nil
	},
}
