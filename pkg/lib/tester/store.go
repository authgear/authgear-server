package tester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type TesterStore struct {
	Context context.Context
	Redis   *globalredis.Handle
}

const TokenLifetime = duration.Short
const ResultLifetime = duration.UserInteraction

func (s *TesterStore) CreateToken(
	appID config.AppID,
	returnURI string,
) (*TesterToken, error) {
	token := NewTesterToken(returnURI)
	bytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		key := redisTokenKey(appID, token.TokenID)
		ttl := TokenLifetime

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

func (s *TesterStore) GetToken(
	appID config.AppID,
	tokenID string,
	consume bool,
) (*TesterToken, error) {
	key := redisTokenKey(appID, tokenID)
	var token TesterToken
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *TesterStore) CreateResult(
	appID config.AppID,
	result *TesterResult,
) (*TesterResult, error) {
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		key := redisResultKey(appID, result.ID)
		ttl := ResultLifetime

		_, err := conn.SetEX(s.Context, key, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *TesterStore) GetResult(
	appID config.AppID,
	resultID string,
) (*TesterResult, error) {
	key := redisResultKey(appID, resultID)
	var result TesterResult
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(s.Context, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrResultNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return err
		}

		return nil
	})
	return &result, err
}

func redisTokenKey(appID config.AppID, tokenID string) string {
	return fmt.Sprintf("app:%s:tester:%s", appID, tokenID)
}

func redisResultKey(appID config.AppID, resultID string) string {
	return fmt.Sprintf("app:%s:tester:result:%s", appID, resultID)
}
