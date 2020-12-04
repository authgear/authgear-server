package oob

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

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
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn redis.Conn) error {
		codeKey := redisCodeKey(s.AppID, code.AuthenticatorID)
		ttl := toMilliseconds(code.ExpireAt.Sub(s.Clock.NowUTC()))
		_, err := goredis.String(conn.Do("SET", codeKey, data, "PX", ttl))
		if errors.Is(err, goredis.ErrNil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) Get(authenticatorID string) (*Code, error) {
	key := redisCodeKey(s.AppID, authenticatorID)
	var codeModel *Code
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", key))
		if errors.Is(err, goredis.ErrNil) {
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
	return s.Redis.WithConn(func(conn redis.Conn) error {
		key := redisCodeKey(s.AppID, authenticatorID)
		_, err := conn.Do("DEL", key)
		if err != nil {
			return err
		}
		return err
	})
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func redisCodeKey(appID config.AppID, authenticatorID string) string {
	return fmt.Sprintf("app:%s:oob-code:%s", appID, authenticatorID)
}
