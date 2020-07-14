package flows

import (
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/redis"
)

type StateStoreRedis struct {
	Redis *redis.Handle
}

func (s *StateStoreRedis) CreateState(state *State) (err error) {
	return s.setState(state, "NX")
}

func (s *StateStoreRedis) UpdateState(state *State) (err error) {
	return s.setState(state, "XX")
}

func (s *StateStoreRedis) setState(state *State, setMode string) (err error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		flowIDKey := flowKey(state.FlowID)
		instanceIDKey := instanceKey(state.InstanceID)
		ttl := toMilliseconds(5 * time.Minute)
		_, err := goredis.String(conn.Do("SET", flowIDKey, []byte(flowIDKey), "PX", ttl, setMode))
		if errors.Is(err, goredis.ErrNil) {
			return errors.New("failed to create interaction flow")
		}
		if err != nil {
			return err
		}

		_, err = goredis.String(conn.Do("SET", instanceIDKey, bytes, "PX", ttl))
		if errors.Is(err, goredis.ErrNil) {
			return errors.New("failed to create interaction flow instance")
		}
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func (s *StateStoreRedis) DeleteState(flowID string) (err error) {
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		flowIDKey := flowKey(flowID)
		_, err := conn.Do("DEL", flowIDKey)
		return err
	})
	return
}

func (s *StateStoreRedis) GetState(instanceID string) (state *State, err error) {
	instanceIDKey := instanceKey(instanceID)
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", instanceIDKey))
		if errors.Is(err, goredis.ErrNil) {
			err = ErrStateNotFound
			return err
		}
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &state)
		if err != nil {
			return err
		}

		flowIDKey := flowKey(state.FlowID)
		_, err = goredis.String(conn.Do("GET", flowIDKey))
		if errors.Is(err, goredis.ErrNil) {
			err = ErrStateNotFound
			return err
		}
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

func flowKey(flowID string) string {
	return fmt.Sprintf("interaction-flow:%s", flowID)
}

func instanceKey(instanceID string) string {
	return fmt.Sprintf("interaction-flow-instance:%s", instanceID)
}
