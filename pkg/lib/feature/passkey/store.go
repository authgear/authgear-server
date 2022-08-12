package passkey

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/duo-labs/webauthn/protocol"
	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type Store struct {
	Context context.Context
	Redis   *appredis.Handle
	AppID   config.AppID
}

func (s *Store) CreateSession(session *Session) error {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(session.Challenge)
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}
	key := redisSessionKey(s.AppID, encodedChallenge)
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		ttl := duration.PerHour
		_, err = conn.SetNX(s.Context, key, bytes, ttl).Result()
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) ConsumeSession(challenge protocol.URLEncodedBase64) (*Session, error) {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(challenge)
	key := redisSessionKey(s.AppID, encodedChallenge)

	var bytes []byte
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		pipeliner := conn.Pipeline()
		getResult := pipeliner.Get(s.Context, key)
		delResult := pipeliner.Del(s.Context, key)
		_, err := pipeliner.Exec(s.Context)
		if err != nil {
			return err
		}
		bytes, err = getResult.Bytes()
		if err != nil {
			return err
		}
		err = delResult.Err()
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

func (s *Store) PeekSession(challenge protocol.URLEncodedBase64) (*Session, error) {
	encodedChallenge := base64.RawURLEncoding.EncodeToString(challenge)
	key := redisSessionKey(s.AppID, encodedChallenge)

	var bytes []byte
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		var err error
		bytes, err = conn.Get(s.Context, key).Bytes()
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
