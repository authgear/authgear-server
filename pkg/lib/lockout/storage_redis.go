package lockout

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type StorageRedis struct {
	AppID config.AppID
	Redis *appredis.Handle
}

var _ Storage = &StorageRedis{}

func (s StorageRedis) Update(spec LockoutSpec, contributor string, delta int) (isSuccess bool, lockedUntil *time.Time, err error) {
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		r, err := makeAttempts(context.Background(), conn,
			redisRecordKey(s.AppID, spec),
			spec.HistoryDuration,
			spec.MaxAttempts,
			spec.MinimumDuration,
			spec.MaximumDuration,
			spec.BackoffFactor,
			spec.IsGlobal,
			contributor,
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

func (s StorageRedis) Clear(spec LockoutSpec, contributor string) (err error) {
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		return clearAttempts(context.Background(), conn,
			redisRecordKey(s.AppID, spec),
			spec.HistoryDuration,
			contributor,
		)
	})
	return err
}

func redisRecordKey(appID config.AppID, spec LockoutSpec) string {
	return fmt.Sprintf("app:%s:lockout:%s", appID, spec.Key())
}
