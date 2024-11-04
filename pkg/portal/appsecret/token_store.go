package appsecret

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type AppSecretVisitTokenStoreImpl struct {
	Redis *globalredis.Handle
}

const Lifetime = duration.Short

func (s *AppSecretVisitTokenStoreImpl) CreateToken(
	ctx context.Context,
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

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisTokenKey(appID, token.TokenID)
		ttl := Lifetime

		_, err := conn.SetEx(ctx, key, bytes, ttl).Result()
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
	ctx context.Context,
	appID config.AppID,
	tokenID string,
) (*AppSecretVisitToken, error) {
	key := redisTokenKey(appID, tokenID)
	var token AppSecretVisitToken
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, key).Bytes()
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
