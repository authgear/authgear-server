package siwe

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

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *StoreRedis) Create(ctx context.Context, nonce *Nonce) error {
	data, err := json.Marshal(nonce)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		nonceKey := redisNonceKey(s.AppID, nonce)
		ttl := nonce.ExpireAt.Sub(s.Clock.NowUTC())
		_, err := conn.SetNX(ctx, nonceKey, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated nonce")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) Get(ctx context.Context, nonce *Nonce) (*Nonce, error) {
	key := redisNonceKey(s.AppID, nonce)
	var nonceModel *Nonce
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrNonceNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &nonceModel)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return nonceModel, nil
}

func (s *StoreRedis) Delete(ctx context.Context, codeKey *Nonce) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisNonceKey(s.AppID, codeKey)
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisNonceKey(appID config.AppID, nonce *Nonce) string {
	return fmt.Sprintf("app:%s:siwe-nonce:%s", appID, nonce.Nonce)
}
