package anonymous

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

var ErrPromotionCodeNotFound = errors.New("promotion code not found")

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *StoreRedis) GetPromotionCode(ctx context.Context, codeHash string) (*PromotionCode, error) {
	c := &PromotionCode{}
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.get(ctx, conn, promotionCodeKey(string(s.AppID), codeHash))
		if err != nil {
			return err
		}
		c, err = s.unmarshalPromotionCode(data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *StoreRedis) CreatePromotionCode(ctx context.Context, code *PromotionCode) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, promotionCodeKey(code.AppID, code.CodeHash), code, code.ExpireAt)
	})
}

// DeletePromotionCode is the point at which a promotion code is spent, so it
// reports ErrPromotionCodeNotFound when there was nothing to delete. DEL is
// atomic, so of several concurrent requests presenting the same code exactly
// one succeeds, which is what stops one anonymous account being linked to
// several targets.
func (s *StoreRedis) DeletePromotionCode(ctx context.Context, code *PromotionCode) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		count, err := conn.Del(ctx, promotionCodeKey(code.AppID, code.CodeHash)).Result()
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrPromotionCodeNotFound
		}
		return nil
	})
}

func (s *StoreRedis) get(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string) ([]byte, error) {
	data, err := conn.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, ErrPromotionCodeNotFound
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *StoreRedis) save(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, value any, expireAt time.Time) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttl := expireAt.Sub(s.Clock.NowUTC())
	_, err = conn.SetNX(ctx, key, data, ttl).Result()
	if errors.Is(err, goredis.Nil) {
		return errors.New("promotion code already exist")
	} else if err != nil {
		return err
	}

	return nil
}

func (s *StoreRedis) unmarshalPromotionCode(data []byte) (*PromotionCode, error) {
	c := PromotionCode{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func promotionCodeKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:promotion-code:%s", appID, tokenHash)
}
