package sso

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type SimpleStoreRedisFactory struct {
	AppID config.AppID
	Redis *appredis.Handle
}

func (f *SimpleStoreRedisFactory) GetStoreByProvider(providerType string, providerAlias string) *SimpleStoreRedis {
	return &SimpleStoreRedis{
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
	redis         *appredis.Handle
	appID         string
	providerType  string
	providerAlias string
}

func (s *SimpleStoreRedis) GetDel(ctx context.Context, key string) (data string, err error) {
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err = s.redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err = conn.Get(ctx, storeKey).Result()
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				return nil
			}
			return err
		}
		_, err = conn.Del(ctx, storeKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *SimpleStoreRedis) SetWithTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err := s.redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.SetEx(ctx, storeKey, []byte(value), ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
