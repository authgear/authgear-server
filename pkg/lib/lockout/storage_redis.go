package lockout

import (
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

func (s StorageRedis) Update(spec BucketSpec, delta int) (lockedUntil *time.Time, err error) {
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		// TODO(tung): TO BE IMPLEMENTED
		return nil
	})
	return nil, err
}

func (s StorageRedis) Clear(spec BucketSpec, delta int) (err error) {
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		// TODO(tung): TO BE IMPLEMENTED
		return nil
	})
	return err
}

func redisBucketKey(appID config.AppID, spec BucketSpec) string {
	return fmt.Sprintf("app:%s:lockout:%s", appID, spec.Key())
}
