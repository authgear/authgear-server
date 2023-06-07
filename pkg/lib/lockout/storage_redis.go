package lockout

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type StorageRedis struct {
	AppID config.AppID
	Redis *appredis.Handle
}

func (s StorageRedis) Update(spec BucketSpec, delta int) (isSuccess bool, lockedUntil *time.Time, err error) {
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		r, err := makeAttempt(context.Background(), conn,
			redisBucketKey(s.AppID, spec),
			spec.HistoryDuration,
			spec.MaxAttempts,
			spec.MinimumDuration,
			spec.MaximumDuration,
			spec.BackoffFactor,
			delta,
		)
		if err != nil {
			return err
		}
		isSuccess = r.IsSuccess
		lockedUntil = r.LockedUntil
		return nil
	})
	return isSuccess, lockedUntil, err
}

func (s StorageRedis) Clear(spec BucketSpec, delta int) (err error) {
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(context.Background(), redisBucketKey(s.AppID, spec)).Result()
		return err
	})
	return err
}

func redisBucketKey(appID config.AppID, spec BucketSpec) string {
	return fmt.Sprintf("app:%s:lockout:%s", appID, spec.Key())
}
