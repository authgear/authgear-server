package otp

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

type CodeStoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

var _ CodeStore = &CodeStoreRedis{}

func (s *CodeStoreRedis) set(ctx context.Context, purpose Purpose, code *Code) error {
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		codeKey := redisCodeKey(s.AppID, purpose, code.Target)
		ttl := code.ExpireAt.Sub(s.Clock.NowUTC())

		_, err := conn.SetEx(ctx, codeKey, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("duplicated code")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *CodeStoreRedis) Create(ctx context.Context, purpose Purpose, code *Code) error {
	return s.set(ctx, purpose, code)
}

func (s *CodeStoreRedis) Get(ctx context.Context, purpose Purpose, target string) (*Code, error) {
	key := redisCodeKey(s.AppID, purpose, target)
	var codeModel *Code
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
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

func (s *CodeStoreRedis) Update(ctx context.Context, purpose Purpose, code *Code) error {
	return s.set(ctx, purpose, code)
}

func (s *CodeStoreRedis) Delete(ctx context.Context, purpose Purpose, target string) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisCodeKey(s.AppID, purpose, target)
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisCodeKey(appID config.AppID, purpose Purpose, target string) string {
	return fmt.Sprintf("app:%s:otp-code:%s:%s", appID, purpose, target)
}
