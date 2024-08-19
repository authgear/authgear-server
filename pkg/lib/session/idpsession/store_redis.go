package idpsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type StoreRedisLogger struct{ *log.Logger }

func NewStoreRedisLogger(lf *log.Factory) StoreRedisLogger {
	return StoreRedisLogger{lf.New("redis-session-store")}
}

type StoreRedis struct {
	Redis  *appredis.Handle
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

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		ctx := context.Background()

		_, err = conn.HSet(ctx, listKey, key, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		_, err = conn.SetNX(ctx, key, json, ttl).Result()
		if errors.Is(err, goredis.Nil) {
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
	ctx := context.Background()
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

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err = conn.HSet(ctx, listKey, key, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		_, err = conn.SetXX(ctx, key, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
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

func (s *StoreRedis) Unmarshal(data []byte) (*IDPSession, error) {
	var sess IDPSession
	err := json.Unmarshal(data, &sess)
	if err != nil {
		return nil, err
	}
	if sess.AuthenticatedAt.IsZero() {
		sess.AuthenticatedAt = sess.CreatedAt
	}

	return &sess, nil
}

func (s *StoreRedis) Get(id string) (*IDPSession, error) {
	ctx := context.Background()
	key := sessionKey(s.AppID, id)

	var sess *IDPSession
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrSessionNotFound
		} else if err != nil {
			return err
		}

		sess, err = s.Unmarshal(data)
		return err
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *StoreRedis) Delete(session *IDPSession) (err error) {
	ctx := context.Background()
	key := sessionKey(s.AppID, session.ID)
	listKey := sessionListKey(s.AppID, session.Attrs.UserID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(ctx, key).Result()
		if err == nil {
			_, err = conn.HDel(ctx, listKey, key).Result()
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

func (s *StoreRedis) CleanUpForDeletingUserID(userID string) (err error) {
	ctx := context.Background()
	listKey := sessionListKey(s.AppID, userID)
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(ctx, listKey).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return
}

//nolint:gocognit
func (s *StoreRedis) List(userID string) (sessions []*IDPSession, err error) {
	ctx := context.Background()
	now := s.Clock.NowUTC()
	listKey := sessionListKey(s.AppID, userID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		sessionList, err := conn.HGetAll(ctx, listKey).Result()
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

			var session *IDPSession
			var sessionJSON []byte
			sessionJSON, err = conn.Get(ctx, key).Bytes()
			// key not found / invalid session JSON -> session not found
			if errors.Is(err, goredis.Nil) {
				err = nil
			} else if err != nil {
				// unexpected error
				return err
			} else {
				session, err = s.Unmarshal(sessionJSON)
				if err != nil {
					s.Logger.
						WithError(err).
						WithFields(logrus.Fields{"key": key}).
						Error("invalid JSON value")
					err = nil
				}
			}

			if session == nil {
				// only cleanup expired sessions from the list
				if expired {
					// ignore non-critical error
					_, err = conn.HDel(ctx, listKey, key).Result()
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

func sessionKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("app:%s:session:%s", appID, sessionID)
}

func sessionListKey(appID config.AppID, userID string) string {
	return fmt.Sprintf("app:%s:session-list:%s", appID, userID)
}
