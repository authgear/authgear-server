package authenticationinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const ttl = duration.Consent

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *StoreRedis) Save(ctx context.Context, entry *Entry) (err error) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return
	}

	key := authenticationInfoEntryKey(s.AppID, entry.ID)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, key, jsonBytes, ttl).Result()
		return err
	})
	if err != nil {
		return
	}

	return
}

func (s *StoreRedis) Get(ctx context.Context, entryID string) (entry *Entry, err error) {
	key := authenticationInfoEntryKey(s.AppID, entryID)
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrNotFound
		} else if err != nil {
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
	return
}

func (s *StoreRedis) Delete(ctx context.Context, entryID string) (err error) {
	key := authenticationInfoEntryKey(s.AppID, entryID)
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, key).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func authenticationInfoEntryKey(appID config.AppID, entryID string) string {
	return fmt.Sprintf("app:%s:authentication-info-entry:%s", appID, entryID)
}
