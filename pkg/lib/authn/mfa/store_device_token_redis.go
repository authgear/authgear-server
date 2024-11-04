package mfa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type StoreDeviceTokenRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

var _ StoreDeviceToken = &StoreDeviceTokenRedis{}

func (s *StoreDeviceTokenRedis) Get(ctx context.Context, userID string, token string) (*DeviceToken, error) {
	var deviceToken *DeviceToken
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisDeviceTokensKey(s.AppID, userID)

		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrDeviceTokenNotFound
		} else if err != nil {
			return err
		}

		tokens := map[string]*DeviceToken{}
		err = json.Unmarshal(data, &tokens)
		if err != nil {
			return err
		}

		if changed, ttl := houseKeepDeviceTokens(tokens, s.Clock.NowUTC()); changed {
			if err := s.saveTokens(ctx, conn, key, tokens, ttl); err != nil {
				return err
			}
		}

		t, ok := tokens[token]
		if !ok {
			return ErrDeviceTokenNotFound
		}

		deviceToken = t
		deviceToken.UserID = userID
		deviceToken.Token = token
		return nil
	})
	if err != nil {
		return nil, err
	}

	return deviceToken, nil
}

func (s *StoreDeviceTokenRedis) Create(ctx context.Context, token *DeviceToken) error {
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisDeviceTokensKey(s.AppID, token.UserID)

		tokens := map[string]*DeviceToken{}
		data, err := conn.Get(ctx, key).Bytes()
		if err != nil {
			if !errors.Is(err, goredis.Nil) {
				return err
			}
		} else {
			if err := json.Unmarshal(data, &tokens); err != nil {
				return err
			}
		}

		if _, exists := tokens[token.Token]; exists {
			return errors.New("duplicated bearer token")
		}
		tokens[token.Token] = token

		_, ttl := houseKeepDeviceTokens(tokens, s.Clock.NowUTC())
		if err := s.saveTokens(ctx, conn, key, tokens, ttl); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *StoreDeviceTokenRedis) DeleteAll(ctx context.Context, userID string) error {
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisDeviceTokensKey(s.AppID, userID)
		_, err := conn.Del(ctx, key).Result()
		return err
	})

	return err
}

func (s *StoreDeviceTokenRedis) HasTokens(ctx context.Context, userID string) (bool, error) {
	count, err := s.Count(ctx, userID)
	if err != nil {
		return false, err
	}
	hasTokens := count > 0
	return hasTokens, nil
}

func (s *StoreDeviceTokenRedis) Count(ctx context.Context, userID string) (int, error) {
	count := 0

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisDeviceTokensKey(s.AppID, userID)
		data, err := conn.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				return nil
			}
			return err
		}

		tokens := map[string]*DeviceToken{}
		if err := json.Unmarshal(data, &tokens); err != nil {
			return err
		}

		count = len(tokens)
		return nil
	})

	return count, err
}

func (s *StoreDeviceTokenRedis) saveTokens(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, tokens map[string]*DeviceToken, ttl time.Duration) error {
	if len(tokens) > 0 {
		data, err := json.Marshal(tokens)
		if err != nil {
			return err
		}

		_, err = conn.Set(ctx, key, data, ttl).Result()
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func houseKeepDeviceTokens(tokens map[string]*DeviceToken, now time.Time) (changed bool, ttl time.Duration) {
	maxExpiry := now
	for token, model := range tokens {
		if now.After(model.ExpireAt) {
			delete(tokens, token)
			changed = true
		} else if model.ExpireAt.After(maxExpiry) {
			maxExpiry = model.ExpireAt
		}
	}

	ttl = maxExpiry.Sub(now)
	return
}

func redisDeviceTokensKey(appID config.AppID, userID string) string {
	return fmt.Sprintf("app:%s:device-tokens:%s", appID, userID)
}
