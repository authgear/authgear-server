package access

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

const maxEventStreamLength = 10

const eventTypeAccessEvent = "access"

type EventStoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *EventStoreRedis) AppendEvent(sessionID string, expiry time.Time, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	streamKey := accessEventStreamKey(s.AppID, sessionID)
	args := &goredis.XAddArgs{
		Stream: streamKey,
		ID:     "*",
		Values: map[string]interface{}{
			eventTypeAccessEvent: data,
		},
	}
	if maxEventStreamLength >= 0 {
		args.MaxLen = maxEventStreamLength
		args.Approx = true
	}

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
		_, err = conn.XAdd(ctx, args).Result()
		if err != nil {
			return err
		}

		_, err = conn.ExpireAt(ctx, streamKey, expiry).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *EventStoreRedis) ResetEventStream(sessionID string) error {
	streamKey := accessEventStreamKey(s.AppID, sessionID)

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
		_, err := conn.Del(ctx, streamKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func accessEventStreamKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("app:%s:access-events:%s", appID, sessionID)
}
