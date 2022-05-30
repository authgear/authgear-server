package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type StoreRedis struct {
	Context context.Context
	Redis   *appredis.Handle
	Clock   clock.Clock
}

func (s *StoreRedis) Create(code *Code) error {
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		codeKey := redisCodeKey(code.Phone)
		ttl := code.ExpireAt.Sub(s.Clock.NowUTC())
		// Using Set to allow overwite whatsapp code
		_, err := conn.Set(s.Context, codeKey, data, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) Update(code *Code) error {
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		codeKey := redisCodeKey(code.Phone)
		ttl := code.ExpireAt.Sub(s.Clock.NowUTC())
		_, err := conn.SetXX(s.Context, codeKey, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrCodeNotFound
		} else if err != nil {
			return err
		}
		return nil
	})
}

func (s *StoreRedis) Get(phone string) (*Code, error) {
	key := redisCodeKey(phone)
	var codeModel *Code
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(s.Context, key).Bytes()
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

func (s *StoreRedis) Delete(phone string) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisCodeKey(phone)
		_, err := conn.Del(s.Context, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisCodeKey(phone string) string {
	return fmt.Sprintf("whatsapp-code:%s", phone)
}
