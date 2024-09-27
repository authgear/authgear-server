package samlslosession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const ttl = duration.UserInteraction

type StoreRedis struct {
	Context context.Context
	Redis   *appredis.Handle
	AppID   config.AppID
}

func (s *StoreRedis) Save(session *SAMLSLOSession) (err error) {
	jsonBytes, err := json.Marshal(session)
	if err != nil {
		return
	}

	key := samlSLOSessionEntryKey(s.AppID, session.ID)

	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(s.Context, key, jsonBytes, ttl).Result()
		return err
	})
	if err != nil {
		return
	}

	return
}

func (s *StoreRedis) Get(sessionID string) (entry *SAMLSLOSession, err error) {
	key := samlSLOSessionEntryKey(s.AppID, sessionID)
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(s.Context, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		var out SAMLSLOSession
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
	key := samlSLOSessionEntryKey(s.AppID, sessionID)
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(s.Context, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func samlSLOSessionEntryKey(appID config.AppID, entryID string) string {
	return fmt.Sprintf("app:%s:saml-slo-session-entry:%s", appID, entryID)
}
