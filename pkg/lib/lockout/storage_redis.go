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

func (s StorageRedis) Update(ctx context.Context, spec LockoutSpec, contributor string, delta int) (isSuccess bool, lockedUntil *time.Time, err error) {
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		r, err := makeAttempts(ctx, conn,
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

func (s StorageRedis) Clear(ctx context.Context, spec LockoutSpec, contributor string) (err error) {
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return clearAttempts(ctx, conn,
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
