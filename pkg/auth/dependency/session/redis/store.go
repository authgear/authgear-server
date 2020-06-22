package redis

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	goredis "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/log"
)

type Store struct {
	Ctx    context.Context
	AppID  string
	Clock  clock.Clock
	Logger *log.Logger
}

var _ session.Store = &Store{}

func (s *Store) Create(sess *session.IDPSession, expireAt time.Time) (err error) {
	json, err := json.Marshal(sess)
	if err != nil {
		return
	}
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Ctx)
	ttl := expireAt.Sub(s.Clock.NowUTC())
	listKey := sessionListKey(s.AppID, sess.Attrs.UserID)
	key := sessionKey(s.AppID, sess.ID)

	_, err = conn.Do("HSET", listKey, key, expiry)
	if err != nil {
		err = errors.Newf("failed to update session list: %w", err)
		return
	}

	_, err = goredis.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "NX"))
	if errors.Is(err, goredis.ErrNil) {
		err = errors.Newf("duplicated session ID: %w", err)
		return
	}

	return
}

func (s *Store) Update(sess *session.IDPSession, expireAt time.Time) (err error) {
	data, err := json.Marshal(sess)
	if err != nil {
		return
	}
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Ctx)
	ttl := expireAt.Sub(s.Clock.NowUTC())
	listKey := sessionListKey(s.AppID, sess.Attrs.UserID)
	key := sessionKey(s.AppID, sess.ID)

	_, err = conn.Do("HSET", listKey, key, expiry)
	if err != nil {
		err = errors.Newf("failed to update session list: %w", err)
		return
	}

	_, err = goredis.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "XX"))
	if errors.Is(err, goredis.ErrNil) {
		err = session.ErrSessionNotFound
	}
	return
}

func (s *Store) Get(id string) (sess *session.IDPSession, err error) {
	conn := redis.GetConn(s.Ctx)
	key := sessionKey(s.AppID, id)
	data, err := goredis.Bytes(conn.Do("GET", key))
	if errors.Is(err, goredis.ErrNil) {
		err = session.ErrSessionNotFound
		return
	} else if err != nil {
		return
	}
	err = json.Unmarshal(data, &sess)
	return
}

func (s *Store) Delete(session *session.IDPSession) (err error) {
	conn := redis.GetConn(s.Ctx)
	key := sessionKey(s.AppID, session.ID)
	listKey := sessionListKey(s.AppID, session.Attrs.UserID)

	_, err = conn.Do("DEL", key)
	if err == nil {
		_, err = conn.Do("HDEL", listKey, key)
		if err != nil {
			s.Logger.
				WithError(err).
				WithField("redis_key", listKey).
				Error("failed to update session list")
			// ignore non-critical errors
			err = nil
		}
	}
	return
}

func (s *Store) List(userID string) (sessions []*session.IDPSession, err error) {
	now := s.Clock.NowUTC()
	conn := redis.GetConn(s.Ctx)
	listKey := sessionListKey(s.AppID, userID)

	sessionList, err := goredis.StringMap(conn.Do("HGETALL", listKey))
	if err != nil {
		return
	}

	for key, expiry := range sessionList {
		expireAt := time.Time{}
		err = expireAt.UnmarshalText([]byte(expiry))
		var expired bool
		if err != nil {
			s.Logger.
				WithError(err).
				WithFields(logrus.Fields{"key": key, "expiry": expiry}).
				Error("invalid expiry value")
			err = nil
			// treat invalid value as expired
			expired = true
		} else {
			expired = now.After(expireAt)
		}

		session := &session.IDPSession{}
		var sessionJSON []byte
		sessionJSON, err = goredis.Bytes(conn.Do("GET", key))
		// key not found / invalid session JSON -> session not found
		if err == goredis.ErrNil {
			err = nil
			session = nil
		} else if err != nil {
			// unexpected error
			return
		} else {
			err = json.Unmarshal(sessionJSON, session)
			if err != nil {
				s.Logger.
					WithError(err).
					WithFields(logrus.Fields{"key": key}).
					Error("invalid JSON value")
				err = nil
				session = nil
			}
		}

		if session == nil {
			// only cleanup expired sessions from the list
			if expired {
				// ignore non-critical error
				_, err = conn.Do("HDEL", listKey, key)
				if err != nil {
					// ignore non-critical error
					s.Logger.
						WithError(err).
						WithFields(logrus.Fields{"key": listKey}).
						Error("failed to update session list")
					err = nil
				}
			}
		} else {
			sessions = append(sessions, session)
		}
	}

	sort.Sort(sessionSlice(sessions))
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

type sessionSlice []*session.IDPSession

func (s sessionSlice) Len() int           { return len(s) }
func (s sessionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionSlice) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }
