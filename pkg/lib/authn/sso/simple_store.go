package sso

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type SimpleStoreRedisFactory struct {
	Context context.Context
	AppID   config.AppID
	Redis   *appredis.Handle
}

func (f *SimpleStoreRedisFactory) GetStoreByProvider(providerType string, providerAlias string) *SimpleStoreRedis {
	return &SimpleStoreRedis{
		context:       f.Context,
		redis:         f.Redis,
		appID:         string(f.AppID),
		providerType:  providerType,
		providerAlias: providerAlias,
	}
}

func storageKey(appID string, providerType string, providerAlias string, key string) string {
	return fmt.Sprintf("app:%s:oauth:%s:%s:%s", appID, providerType, providerAlias, key)
}

type SimpleStoreRedis struct {
	context       context.Context
	redis         *appredis.Handle
	appID         string
	providerType  string
	providerAlias string
}

func (s *SimpleStoreRedis) GetDel(key string) (data string, err error) {
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err = s.redis.WithConnContext(s.context, func(conn *goredis.Conn) error {
		data, err = conn.Get(s.context, storeKey).Result()
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				return nil
			}
			return err
		}
		_, err = conn.Del(s.context, storeKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *SimpleStoreRedis) SetWithTTL(key string, value string, ttl time.Duration) error {
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err := s.redis.WithConnContext(s.context, func(conn *goredis.Conn) error {
		_, err := conn.SetEX(s.context, storeKey, []byte(value), ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
