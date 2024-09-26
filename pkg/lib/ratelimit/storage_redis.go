package ratelimit

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

func (s StorageRedis) Update(spec BucketSpec, delta int) (ok bool, timeToAct time.Time, err error) {
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		result, err := gcra(context.Background(), conn,
			redisBucketKey(s.AppID, spec), spec.Period, spec.Burst, delta,
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

func redisBucketKey(appID config.AppID, spec BucketSpec) string {
	if spec.IsGlobal {
		return fmt.Sprintf("rate-limit:%s", spec.Key())
	}
	return fmt.Sprintf("app:%s:rate-limit:%s", appID, spec.Key())
}
