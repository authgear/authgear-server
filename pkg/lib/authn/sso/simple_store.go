package sso

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type SimpleStoreRedisFactory struct {
	Redis *appredis.Handle
}

func (f *SimpleStoreRedisFactory) GetStoreByOAuthType(typ string) *SimpleStoreRedis {
	return &SimpleStoreRedis{
		redis:     f.Redis,
		oauthType: typ,
	}
}

func storageKey(oauthType string, key string) string {
	return fmt.Sprint("oauth:%s:simple-store:key", oauthType, key)
}

type SimpleStoreRedis struct {
	redis     *appredis.Handle
	oauthType string
}

func (s *SimpleStoreRedis) GetDel(key string) (data string, err error) {
	ctx := context.Background()
	storeKey := storageKey(s.oauthType, key)
	err = s.redis.WithConn(func(conn *goredis.Conn) error {
		data, err = conn.GetDel(ctx, storeKey).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (s *SimpleStoreRedis) SetWithTTL(key string, value string, ttl time.Duration) error {
	ctx := context.Background()
	storeKey := storageKey(s.oauthType, key)
	err := s.redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.SetEX(ctx, storeKey, []byte(value), ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
