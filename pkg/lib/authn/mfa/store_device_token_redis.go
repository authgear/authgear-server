package mfa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

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

func (s *StoreDeviceTokenRedis) Get(userID string, token string) (*DeviceToken, error) {
	var deviceToken *DeviceToken
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
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
			if err := s.saveTokens(conn, key, tokens, ttl); err != nil {
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

func (s *StoreDeviceTokenRedis) Create(token *DeviceToken) error {
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
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
		if err := s.saveTokens(conn, key, tokens, ttl); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *StoreDeviceTokenRedis) DeleteAll(userID string) error {
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		key := redisDeviceTokensKey(s.AppID, userID)
		ctx := context.Background()
		_, err := conn.Del(ctx, key).Result()
		return err
	})

	return err
}

func (s *StoreDeviceTokenRedis) HasTokens(userID string) (bool, error) {
	hasTokens := false

	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		ctx := context.Background()
		key := redisDeviceTokensKey(s.AppID, userID)
		_, err := conn.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				hasTokens = false
				return nil
			}
			return err
		}
		hasTokens = true
		return nil
	})

	return hasTokens, err
}

func (s *StoreDeviceTokenRedis) saveTokens(conn redis.Redis_6_0_Cmdable, key string, tokens map[string]*DeviceToken, ttl time.Duration) error {
	ctx := context.Background()

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
