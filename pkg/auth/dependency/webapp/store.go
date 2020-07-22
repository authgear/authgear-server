package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/redis"
)

var ErrInvalidState = errors.New("webapp-store: invalid state")

type RedisStore struct {
	AppID config.AppID
	Redis *redis.Handle
}

func (s *RedisStore) Create(state *State) (err error) {
	keyFlow := flowKey(string(s.AppID), state.FlowID)
	keyInstance := instanceKey(string(s.AppID), state.InstanceID)
	bytes, err := json.Marshal(state)
	if err != nil {
		return
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		ttl := toMilliseconds(5 * time.Minute)
		_, err := goredis.String(conn.Do("SET", keyFlow, []byte(keyFlow), "PX", ttl))
		if err != nil {
			return err
		}
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

		keyFlow := flowKey(string(s.AppID), state.FlowID)
		_, err = goredis.String(conn.Do("GET", keyFlow))
		if errors.Is(err, goredis.ErrNil) {
			err = ErrInvalidState
			return err
		}
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *RedisStore) DeleteFlow(flowID string) (err error) {
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		keyFlow := flowKey(string(s.AppID), flowID)
		_, err := conn.Do("DEL", keyFlow)
		return err
	})
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func flowKey(appID string, flowID string) string {
	return fmt.Sprintf("app:%s:webapp-state-flow:%s", appID, flowID)
}

func instanceKey(appID string, instanceID string) string {
	return fmt.Sprintf("app:%s:webapp-state-instance:%s", appID, instanceID)
}
