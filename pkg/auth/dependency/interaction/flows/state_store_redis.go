package flows

import (
	"encoding/json"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/redis"
)

type StateStoreRedis struct {
	Redis *redis.Handle
}

func (s *StateStoreRedis) Get(id string) (state *State, err error) {
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", id))
		if errors.Is(err, goredis.ErrNil) {
			err = ErrStateNotFound
			return err
		} else if err != nil {
			return err
		}
		err = json.Unmarshal(data, &state)
		return err
	})
	return
}

func (s *StateStoreRedis) Set(state *State) (err error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = goredis.String(conn.Do("SET", state.ID, bytes, "PX", toMilliseconds(5*time.Minute)))
		return err
	})

	return
}

func (s *StateStoreRedis) Delete(id string) (err error) {
	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := conn.Do("DEL", id)
		return err
	})
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}
