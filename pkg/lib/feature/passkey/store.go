package passkey

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type Store struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *Store) CreateSession(ctx context.Context, session *Session) error {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(session.Challenge)
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := redisSessionKey(s.AppID, encodedChallenge)
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		ttl := duration.PerHour
		_, err = conn.SetNX(ctx, key, bytes, ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) ConsumeSession(ctx context.Context, challenge protocol.URLEncodedBase64) (*Session, error) {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(challenge)
	key := redisSessionKey(s.AppID, encodedChallenge)

	var bytes []byte
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		bytes, err = conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrSessionNotFound
		}
		if err != nil {
			return err
		}

		_, err = conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(bytes, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *Store) PeekSession(ctx context.Context, challenge protocol.URLEncodedBase64) (*Session, error) {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(challenge)
	key := redisSessionKey(s.AppID, encodedChallenge)

	var bytes []byte
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		bytes, err = conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrSessionNotFound
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(bytes, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func redisSessionKey(appID config.AppID, encodedChallenge string) string {
	return fmt.Sprintf("app:%s:passkey-session:%s", appID, encodedChallenge)
}
