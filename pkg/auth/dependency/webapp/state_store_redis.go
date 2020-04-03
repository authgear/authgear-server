package webapp

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

type StateStoreImpl struct {
	Context context.Context
}

var _ StateStore = &StateStoreImpl{}

func (s *StateStoreImpl) Get(id string) (state *State, err error) {
	conn := redis.GetConn(s.Context)
	data, err := goredis.Bytes(conn.Do("GET", id))
	if errors.Is(err, goredis.ErrNil) {
		err = ErrStateNotFound
		return
	} else if err != nil {
		return
	}
	err = json.Unmarshal(data, &state)
	return
}

func (s *StateStoreImpl) Set(state *State) (err error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Context)
	_, err = goredis.String(conn.Do("SET", state.ID, bytes, "PX", toMilliseconds(5*time.Minute)))
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}
