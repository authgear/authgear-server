package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

var ErrInvalidState = apierrors.Invalid.WithReason("WebUIInvalidState").New("the claimed session is invalid")

type RedisStore struct {
	AppID config.AppID
	Redis *redis.Handle
}

func (s *RedisStore) Create(state *State) (err error) {
	keyInstance := instanceKey(string(s.AppID), state.ID)
	bytes, err := json.Marshal(state)
	if err != nil {
		return
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		ttl := toMilliseconds(5 * time.Minute)
		_, err = goredis.String(conn.Do("SET", keyInstance, bytes, "PX", ttl, "NX"))
		if errors.Is(err, goredis.ErrNil) {
			return fmt.Errorf("webapp-store: failed to create state: %w", err)
		}
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (s *RedisStore) Get(instanceID string) (state *State, err error) {
	instanceKey := instanceKey(string(s.AppID), instanceID)
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", instanceKey))
		if errors.Is(err, goredis.ErrNil) {
			err = ErrInvalidState
			return err
		}
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &state)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func instanceKey(appID string, instanceID string) string {
	return fmt.Sprintf("app:%s:webapp-state-instance:%s", appID, instanceID)
}
