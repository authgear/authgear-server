package samlsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const ttl = duration.UserInteraction + duration.Consent

type StoreRedis struct {
	Context context.Context
	Redis   *appredis.Handle
	AppID   config.AppID
}

func (s *StoreRedis) Save(session *SAMLSession) (err error) {
	jsonBytes, err := json.Marshal(session)
	if err != nil {
		return
	}

	key := samlSessionEntryKey(s.AppID, session.ID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Set(s.Context, key, jsonBytes, ttl).Result()
		return err
	})
	if err != nil {
		return
	}

	return
}

func (s *StoreRedis) Get(sessionID string) (entry *SAMLSession, err error) {
	key := samlSessionEntryKey(s.AppID, sessionID)
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(s.Context, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		var out SAMLSession
		err = json.Unmarshal(data, &out)
		if err != nil {
			return err
		}
		entry = &out
		return nil
	})
	return
}

func (s *StoreRedis) Delete(sessionID string) (err error) {
	key := samlSessionEntryKey(s.AppID, sessionID)
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(s.Context, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func samlSessionEntryKey(appID config.AppID, entryID string) string {
	return fmt.Sprintf("app:%s:saml-session-entry:%s", appID, entryID)
}
