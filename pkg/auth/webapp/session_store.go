package webapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
)

const SessionExpiryDuration = interaction.GraphLifetime

type SessionStoreRedis struct {
	AppID config.AppID
	Redis *appredis.Handle
}

func (s *SessionStoreRedis) Create(ctx context.Context, session *Session) (err error) {
	key := sessionKey(string(s.AppID), session.ID)
	bytes, err := json.Marshal(session)
	if err != nil {
		return
	}

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		ttl := SessionExpiryDuration
		_, err = conn.SetNX(ctx, key, bytes, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return fmt.Errorf("webapp-store: failed to create session: %w", err)
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	otelauthgear.IntCounterAddOne(
		ctx,
		otelauthgear.CounterWebSessionCreationCount,
	)

	return
}

func (s *SessionStoreRedis) Update(ctx context.Context, session *Session) (err error) {
	key := sessionKey(string(s.AppID), session.ID)
	bytes, err := json.Marshal(session)
	if err != nil {
		return
	}

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		ttl := SessionExpiryDuration
		_, err = conn.SetXX(ctx, key, bytes, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrInvalidSession
		}
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (s *SessionStoreRedis) Get(ctx context.Context, id string) (session *Session, err error) {
	key := sessionKey(string(s.AppID), id)
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			err = ErrInvalidSession
			return err
		}
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &session)
		// translation logic end
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *SessionStoreRedis) Delete(ctx context.Context, id string) error {
	key := sessionKey(string(s.AppID), id)
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, key).Result()
		return err
	})
	return err
}

func sessionKey(appID string, id string) string {
	return fmt.Sprintf("app:%s:webapp-session:%s", appID, id)
}
