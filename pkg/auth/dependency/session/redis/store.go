package redis

import (
	"encoding/json"
	"sort"
	"time"

	goredis "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errors"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("redis-session-store")} }

type Store struct {
	Redis  *redis.Handle
	AppID  config.AppID
	Clock  clock.Clock
	Logger Logger
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

	ttl := expireAt.Sub(s.Clock.NowUTC())
	listKey := sessionListKey(s.AppID, sess.Attrs.UserID)
	key := sessionKey(s.AppID, sess.ID)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = conn.Do("HSET", listKey, key, expiry)
		if err != nil {
			return errors.Newf("failed to update session list: %w", err)
		}

		_, err = goredis.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "NX"))
		if errors.Is(err, goredis.ErrNil) {
			err = errors.Newf("duplicated session ID: %w", err)
			return err
		}

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
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

	ttl := expireAt.Sub(s.Clock.NowUTC())
	listKey := sessionListKey(s.AppID, sess.Attrs.UserID)
	key := sessionKey(s.AppID, sess.ID)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = conn.Do("HSET", listKey, key, expiry)
		if err != nil {
			return errors.Newf("failed to update session list: %w", err)
		}

		_, err = goredis.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "XX"))
		if errors.Is(err, goredis.ErrNil) {
			return session.ErrSessionNotFound
		}

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return
	}

	return
}

func (s *Store) Get(id string) (sess *session.IDPSession, err error) {
	key := sessionKey(s.AppID, id)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", key))
		if errors.Is(err, goredis.ErrNil) {
			return session.ErrSessionNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &sess)
		return err
	})
	return
}

func (s *Store) Delete(session *session.IDPSession) (err error) {
	key := sessionKey(s.AppID, session.ID)
	listKey := sessionListKey(s.AppID, session.Attrs.UserID)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := conn.Do("DEL", key)
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
		return err
	})
	return
}

func (s *Store) List(userID string) (sessions []*session.IDPSession, err error) {
	now := s.Clock.NowUTC()
	listKey := sessionListKey(s.AppID, userID)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		sessionList, err := goredis.StringMap(conn.Do("HGETALL", listKey))
		if err != nil {
			return err
		}

		for key, expiry := range sessionList {
			expireAt := time.Time{}
			err = expireAt.UnmarshalText([]byte(expiry))
			var expired bool
			if err != nil {
				err = nil
				s.Logger.
					WithError(err).
					WithFields(logrus.Fields{"key": key, "expiry": expiry}).
					Error("invalid expiry value")
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
				return err
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
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return
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
