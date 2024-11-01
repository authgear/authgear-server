package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type TokenStore struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *TokenStore) Set(ctx context.Context, token *UserToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisTokenKey(s.AppID, token.Endpoint, token.Username)
		ttl := token.ExpireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetEx(ctx, key, data, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *TokenStore) Get(ctx context.Context, endpoint string, username string) (*UserToken, error) {
	key := redisTokenKey(s.AppID, endpoint, username)
	var token *UserToken
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return nil
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &token)
		if err != nil {
			return err
		}

		return nil
	})
	return token, err
}

func redisTokenKey(appID config.AppID, endpoint string, username string) string {
	return fmt.Sprintf("app:%s:whatsapp-on-prem-token:%s:%s", appID, endpoint, username)
}
