package challenge

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type Store struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *Store) Save(ctx context.Context, c *Challenge, ttl time.Duration) error {
	key := challengeKey(s.AppID, c.Token)
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err = conn.SetNX(ctx, key, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("fail to create new challenge")
		} else if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Get(ctx context.Context, token string, consume bool) (*Challenge, error) {
	key := challengeKey(s.AppID, token)

	c := &Challenge{}

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrInvalidChallenge
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, c)
		if err != nil {
			return err
		}

		if consume {
			_, err = conn.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
