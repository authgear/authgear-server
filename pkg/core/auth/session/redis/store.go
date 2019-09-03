package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	gotime "time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type store struct {
	ctx   context.Context
	appID string
	time  time.Provider
}

var _ session.Store = &store{}

var errSessionCreateFailed = fmt.Errorf("cannot create session")

func NewStore(ctx context.Context, appID string, time time.Provider) session.Store {
	return &store{ctx: ctx, appID: appID, time: time}
}

func (s *store) Create(sess *auth.Session, expireAt gotime.Time) (err error) {
	json, err := json.Marshal(sess)
	if err != nil {
		return
	}
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return
	}

	conn := redis.GetConn(s.ctx)
	ttl := expireAt.Sub(s.time.NowUTC())
	listKey := sessionListKey(s.appID, sess.UserID)
	key := sessionKey(s.appID, sess.ID)

	_, err = conn.Do("HSET", listKey, key, expiry)
	if err != nil {
		err = fmt.Errorf("cannot create session")
		return
	}

	_, err = goredis.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "NX"))
	if err == goredis.ErrNil {
		err = errSessionCreateFailed
		return
	}

	return
}

func (s *store) Update(sess *auth.Session, expireAt gotime.Time) (err error) {
	data, err := json.Marshal(sess)
	if err != nil {
		return
	}
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return
	}

	conn := redis.GetConn(s.ctx)
	ttl := expireAt.Sub(s.time.NowUTC())
	listKey := sessionListKey(s.appID, sess.UserID)
	key := sessionKey(s.appID, sess.ID)

	_, err = conn.Do("HSET", listKey, key, expiry)
	if err != nil {
		err = fmt.Errorf("cannot create session")
		return
	}

	_, err = goredis.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "XX"))
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

func (s *store) Delete(session *auth.Session) (err error) {
	conn := redis.GetConn(s.ctx)
	key := sessionKey(s.appID, session.ID)
	listKey := sessionListKey(s.appID, session.UserID)

	_, err = conn.Do("DEL", key)
	// ignore non-critical error
	_, _ = conn.Do("HDEL", listKey, key)
	return
}

func (s *store) DeleteBatch(sessions []*auth.Session) (err error) {
	conn := redis.GetConn(s.ctx)

	sessionKeys := []interface{}{}
	listKeys := map[string]struct{}{}
	for _, session := range sessions {
		sessionKeys = append(sessionKeys, sessionKey(s.appID, session.ID))
		listKeys[sessionListKey(s.appID, session.UserID)] = struct{}{}
	}
	_, err = conn.Do("DEL", sessionKeys...)

	for listKey := range listKeys {
		// ignore non-critical error
		_, _ = conn.Do("HDEL", append([]interface{}{listKey}, sessionKeys...))
	}
	return
}

func (s *store) DeleteAll(userID string, sessionID string) error {
	conn := redis.GetConn(s.ctx)
	listKey := sessionListKey(s.appID, userID)

	sessionKeys, err := goredis.Strings(conn.Do("HKEYS", listKey))
	if err != nil {
		return err
	}

	sessionKeysDel := []interface{}{}
	excludeSessionkey := sessionKey(s.appID, sessionID)
	for _, sessionKey := range sessionKeys {
		if excludeSessionkey == sessionKey {
			continue
		}
		sessionKeysDel = append(sessionKeysDel, sessionKey)
	}
	if len(sessionKeysDel) == 0 {
		return nil
	}

	_, err = conn.Do("DEL", sessionKeysDel...)
	if err != nil {
		return err
	}

	args := append([]interface{}{listKey}, sessionKeysDel...)
	// ignore non-critical error
	_, _ = conn.Do("HDEL", args...)

	return nil
}

func (s *store) List(userID string) (sessions []*auth.Session, err error) {
	now := s.time.NowUTC()
	conn := redis.GetConn(s.ctx)
	listKey := sessionListKey(s.appID, userID)

	sessionList, err := goredis.StringMap(conn.Do("HGETALL", listKey))
	if err != nil {
		return
	}

	for key, expiry := range sessionList {
		expireAt := gotime.Time{}
		err = expireAt.UnmarshalText([]byte(expiry))
		var expired bool
		if err != nil {
			err = nil
			// treat invalid value as expired
			expired = true
		} else {
			expired = now.After(expireAt)
		}

		session := &auth.Session{}
		var sessionJSON []byte
		sessionJSON, err = goredis.Bytes(conn.Do("GET", key))
		// key not found / invalid session JSON -> session not found
		if err == goredis.ErrNil {
			err = nil
			session = nil
		} else {
			err = json.Unmarshal(sessionJSON, session)
			if err != nil {
				err = nil
				session = nil
			}
		}

		if session == nil {
			// only cleanup expired sessions from the list
			if expired {
				// ignore non-critical error
				_, _ = conn.Do("HDEL", listKey, key)
			}
		} else {
			sessions = append(sessions, session)
		}
	}

	sort.Sort(sessionSlice(sessions))
	return
}

func toMilliseconds(d gotime.Duration) int64 {
	return int64(d / gotime.Millisecond)
}

type sessionSlice []*auth.Session

func (s sessionSlice) Len() int           { return len(s) }
func (s sessionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionSlice) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }
