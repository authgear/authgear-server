package redis

import (
	"context"
	"encoding/json"
	"fmt"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

type store struct {
	ctx   context.Context
	appID string
}

var _ session.Store = &store{}

var errSessionCreateFailed = fmt.Errorf("cannot create session")

func NewStore(ctx context.Context, appID string) session.Store {
	return &store{ctx: ctx, appID: appID}
}

func (s *store) Create(sess *auth.Session) (err error) {
	json, err := json.Marshal(sess)
	if err != nil {
		return
	}
	conn := redis.GetConn(s.ctx)
	key := sessionKey(s.appID, sess.ID)
	_, err = goredis.String(conn.Do("SET", key, json, "NX"))
	if err == goredis.ErrNil {
		err = errSessionCreateFailed
	}
	return
}

func (s *store) Update(sess *auth.Session) (err error) {
	data, err := json.Marshal(sess)
	if err != nil {
		return
	}
	conn := redis.GetConn(s.ctx)
	key := sessionKey(s.appID, sess.ID)
	_, err = goredis.String(conn.Do("SET", key, data, "XX"))
	if err == goredis.ErrNil {
		err = session.ErrSessionNotFound
	}
	return
}

func (s *store) Get(id string) (sess *auth.Session, err error) {
	conn := redis.GetConn(s.ctx)
	key := sessionKey(s.appID, id)
	data, err := goredis.Bytes(conn.Do("GET", key))
	if err == goredis.ErrNil {
		err = session.ErrSessionNotFound
		return
	}
	err = json.Unmarshal(data, &sess)
	return
}

func (s *store) Delete(id string) (err error) {
	conn := redis.GetConn(s.ctx)
	key := sessionKey(s.appID, id)
	_, err = conn.Do("DEL", key)
	return
}
