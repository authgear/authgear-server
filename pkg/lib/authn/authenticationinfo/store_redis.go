package authenticationinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const ttl = duration.Short

type StoreRedis struct {
	Context context.Context
	Redis   *redis.Handle
	AppID   config.AppID
}

func (s *StoreRedis) Save(entry *Entry) (err error) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return
	}

	key := authenticationInfoEntryKey(s.AppID, entry.ID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Set(s.Context, key, jsonBytes, ttl).Result()
		return err
	})
	if err != nil {
		return
	}

	return
}

func (s *StoreRedis) Consume(entryID string) (entry *Entry, err error) {
	key := authenticationInfoEntryKey(s.AppID, entryID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		pipeline := conn.TxPipeline()
		get := pipeline.Get(s.Context, key)
		del := pipeline.Del(s.Context, key)

		_, err := pipeline.Exec(s.Context)
		if err != nil {
			return err
		}

		data, err := get.Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrNotFound
		} else if err != nil {
			return err
		}

		_, err = del.Result()
		if err != nil {
			return err
		}

		var out Entry
		err = json.Unmarshal(data, &out)
		if err != nil {
			return err
		}

		entry = &out

		return nil
	})
	if err != nil {
		return
	}

	return
}

func authenticationInfoEntryKey(appID config.AppID, entryID string) string {
	return fmt.Sprintf("app:%s:authentication-info-entry:%s", appID, entryID)
}
