package appsecret

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

type AppSecretVisitTokenStoreImpl struct {
	Context context.Context
	Redis   *globalredis.Handle
}

const Lifetime = duration.Short

func (s *AppSecretVisitTokenStoreImpl) CreateToken(
	appID config.AppID,
	userID string,
	secrets []config.SecretKey,
) (*AppSecretVisitToken, error) {
	token := NewAppSecretVisitToken(
		userID,
		secrets,
	)
	bytes, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *AppSecretVisitTokenStoreImpl) GetTokenByID(
	appID config.AppID,
	tokenID string,
) (*AppSecretVisitToken, error) {
	key := redisTokenKey(appID, tokenID)
	var token AppSecretVisitToken
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

		return nil
	})
	return &token, err
}

func redisTokenKey(appID config.AppID, tokenID string) string {
	return fmt.Sprintf("app:%s:secret-visit-token:%s", appID, tokenID)
}
