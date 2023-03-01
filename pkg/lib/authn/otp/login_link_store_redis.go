package otp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type LoginLinkStoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *LoginLinkStoreRedis) set(token string, target string, expireAt time.Time) error {
	ctx := context.Background()
	data, err := json.Marshal(target)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisLoginLinkKey(s.AppID, token)
		ttl := expireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetNX(ctx, key, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *LoginLinkStoreRedis) Create(token string, target string, expireAt time.Time) error {
	return s.set(token, target, expireAt)
}

func (s *LoginLinkStoreRedis) Get(token string) (string, error) {
	ctx := context.Background()
	key := redisLoginLinkKey(s.AppID, token)
	var target string
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrCodeNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &target)
		if err != nil {
			return err
		}

		return nil
	})
	return target, err
}

func (s *LoginLinkStoreRedis) Delete(token string) error {
	ctx := context.Background()
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisLoginLinkKey(s.AppID, token)
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisLoginLinkKey(appID config.AppID, token string) string {
	return fmt.Sprintf("app:%s:login-link:%s", appID, token)
}
