package ratelimit

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
)

type StorageRedis struct {
	Redis *redis.Handle
}

func NewAppStorageRedis(redis *appredis.Handle) *StorageRedis {
	return &StorageRedis{Redis: redis.Handle}
}

func NewGlobalStorageRedis(redis *globalredis.Handle) *StorageRedis {
	return &StorageRedis{Redis: redis.Handle}
}

func (s *StorageRedis) Update(key string, period time.Duration, burst int, delta int) (ok bool, timeToAct time.Time, err error) {
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		result, err := gcra(context.Background(), conn,
			key, period, burst, delta,
		)
		if err != nil {
			return err
		}
		ok = result.IsConforming
		timeToAct = result.TimeToAct
		return nil
	})
	return
}
