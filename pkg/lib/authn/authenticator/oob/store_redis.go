package oob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type StoreRedis struct {
	Redis *redis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *StoreRedis) Create(code *Code) error {
	ctx := context.Background()
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		codeKey := redisCodeKey(s.AppID, code.AuthenticatorID)
		ttl := code.ExpireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetNX(ctx, codeKey, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) Get(authenticatorID string) (*Code, error) {
	ctx := context.Background()
	key := redisCodeKey(s.AppID, authenticatorID)
	var codeModel *Code
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrCodeNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &codeModel)
		if err != nil {
			return err
		}

		return nil
	})
	return codeModel, err
}

func (s *StoreRedis) Delete(authenticatorID string) error {
	ctx := context.Background()
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisCodeKey(s.AppID, authenticatorID)
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisCodeKey(appID config.AppID, authenticatorID string) string {
	return fmt.Sprintf("app:%s:oob-code:%s", appID, authenticatorID)
}
