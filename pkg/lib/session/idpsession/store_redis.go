package idpsession

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	goredis "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type StoreRedisLogger struct{ *log.Logger }

func NewStoreRedisLogger(lf *log.Factory) StoreRedisLogger {
	return StoreRedisLogger{lf.New("redis-session-store")}
}

type StoreRedis struct {
	Redis  *redis.Handle
	AppID  config.AppID
	Clock  clock.Clock
	Logger StoreRedisLogger
}

func (s *StoreRedis) Create(sess *IDPSession, expireAt time.Time) (err error) {
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
			return fmt.Errorf("failed to update session list: %w", err)
		}

		_, err = goredis.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "NX"))
		if errorutil.Is(err, goredis.ErrNil) {
			err = fmt.Errorf("duplicated session ID: %w", err)
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

func (s *StoreRedis) Update(sess *IDPSession, expireAt time.Time) (err error) {
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
			return fmt.Errorf("failed to update session list: %w", err)
		}

		_, err = goredis.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "XX"))
		if errorutil.Is(err, goredis.ErrNil) {
			return ErrSessionNotFound
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

func (s *StoreRedis) Get(id string) (sess *IDPSession, err error) {
	key := sessionKey(s.AppID, id)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", key))
		if errorutil.Is(err, goredis.ErrNil) {
			return ErrSessionNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &sess)
		return err
	})
	return
}

func (s *StoreRedis) Delete(session *IDPSession) (err error) {
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

func (s *StoreRedis) List(userID string) (sessions []*IDPSession, err error) {
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

			session := &IDPSession{}
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

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
	})
	return
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func sessionKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("app:%s:session:%s", appID, sessionID)
}

func sessionListKey(appID config.AppID, userID string) string {
	return fmt.Sprintf("app:%s:session-list:%s", appID, userID)
}
