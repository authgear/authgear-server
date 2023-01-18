package otp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type CodeStoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *CodeStoreRedis) set(target string, code *Code) error {
	ctx := context.Background()
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		codeKey := redisCodeKey(s.AppID, target)
		ttl := code.ExpireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetEX(ctx, codeKey, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *CodeStoreRedis) Create(target string, code *Code) error {
	return s.set(target, code)
}

func (s *CodeStoreRedis) Get(target string) (*Code, error) {
	ctx := context.Background()
	key := redisCodeKey(s.AppID, target)
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

func (s *CodeStoreRedis) Update(target string, code *Code) error {
	return s.set(target, code)
}

func (s *CodeStoreRedis) Delete(target string) error {
	ctx := context.Background()
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisCodeKey(s.AppID, target)
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisCodeKey(appID config.AppID, target string) string {
	return fmt.Sprintf("app:%s:otp-code:%s", appID, target)
}
