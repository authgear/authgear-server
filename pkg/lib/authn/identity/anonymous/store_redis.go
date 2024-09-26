package anonymous

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

var ErrPromotionCodeNotFound = errors.New("promotion code not found")

type StoreRedis struct {
	Context context.Context
	Redis   *appredis.Handle
	AppID   config.AppID
	Clock   clock.Clock
}

func (s *StoreRedis) GetPromotionCode(codeHash string) (*PromotionCode, error) {
	c := &PromotionCode{}
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		data, err := s.get(conn, promotionCodeKey(string(s.AppID), codeHash))
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

func (s *StoreRedis) CreatePromotionCode(code *PromotionCode) error {
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		return s.save(conn, promotionCodeKey(code.AppID, code.CodeHash), code, code.ExpireAt)
	})
}

func (s *StoreRedis) DeletePromotionCode(code *PromotionCode) error {
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		return s.del(conn, promotionCodeKey(code.AppID, code.CodeHash))
	})
}

func (s *StoreRedis) get(conn redis.Redis_6_0_Cmdable, key string) ([]byte, error) {
	ctx := context.Background()
	data, err := conn.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, ErrPromotionCodeNotFound
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *StoreRedis) save(conn redis.Redis_6_0_Cmdable, key string, value interface{}, expireAt time.Time) error {
	ctx := context.Background()
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

func (s *StoreRedis) del(conn redis.Redis_6_0_Cmdable, key string) error {
	ctx := context.Background()
	_, err := conn.Del(ctx, key).Result()
	return err
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
