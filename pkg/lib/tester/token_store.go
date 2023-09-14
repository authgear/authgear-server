package tester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type TesterTokenStore struct {
	Context context.Context
	Redis   *globalredis.Handle
}

const Lifetime = duration.Short

func (s *TesterTokenStore) CreateToken(
	appID config.AppID,
	returnURI string,
) (*TesterToken, error) {
	token := NewTesterToken(returnURI)
	bytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := redisTokenKey(appID, token.TokenID)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, key, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *TesterTokenStore) GetToken(
	appID config.AppID,
	tokenID string,
	consume bool,
) (*TesterToken, error) {
	key := redisTokenKey(appID, tokenID)
	var token TesterToken
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		bytes, err := conn.Get(s.Context, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrTokenNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &token)
		if err != nil {
			return err
		}

		if consume {
			_, err = conn.Del(s.Context, key).Result()
			if err != nil {
				return err
			}
		}

		return nil
	})
	return &token, err
}

func redisTokenKey(appID config.AppID, tokenID string) string {
	return fmt.Sprintf("app:%s:tester:%s", appID, tokenID)
}
