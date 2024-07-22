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

func (s *SimpleStoreRedis) GetDel(key string) (data string, err error) {
	ctx := context.Background()
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err = s.redis.WithConn(func(conn *goredis.Conn) error {
		data, err = conn.GetDel(ctx, storeKey).Result()
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				return nil
			}
			return err
		}
		return nil
	})
	return
}

func (s *SimpleStoreRedis) SetWithTTL(key string, value string, ttl time.Duration) error {
	ctx := context.Background()
	storeKey := storageKey(s.appID, s.providerType, s.providerAlias, key)
	err := s.redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.SetEX(ctx, storeKey, []byte(value), ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
