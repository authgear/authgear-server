package idpsession

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var StoreRedisLogger = slogutil.NewLogger("redis-session-store")

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *StoreRedis) Create(ctx context.Context, sess *IDPSession, expireAt time.Time) (err error) {
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

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
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

	logger := StoreRedisLogger.GetLogger(ctx)
	// NOTE(DEV-2982): This is for debugging the session lost problem
	logger.WithSkipStackTrace().Warn(ctx,
		"create IDP session",
		slog.String("idp_session_id", sess.ID),
		slog.Time("idp_session_created_at", sess.CreatedAt),
		slog.String("user_id", sess.Attrs.UserID),
		slog.Bool("refresh_token_log", true),
	)

	return
}

func (s *StoreRedis) Update(ctx context.Context, sess *IDPSession, expireAt time.Time) (err error) {
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

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
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

func (s *StoreRedis) Get(ctx context.Context, id string) (*IDPSession, error) {
	key := sessionKey(s.AppID, id)

	var sess *IDPSession
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
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

func (s *StoreRedis) Delete(ctx context.Context, session *IDPSession) (err error) {
	logger := StoreRedisLogger.GetLogger(ctx)
	key := sessionKey(s.AppID, session.ID)
	listKey := sessionListKey(s.AppID, session.Attrs.UserID)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, key).Result()
		if err == nil {
			_, err = conn.HDel(ctx, listKey, key).Result()
			if err != nil {
				logger.WithError(err).Error(ctx, "failed to update session list",
					slog.String("redis_key", listKey),
				)
				// ignore non-critical errors
				err = nil
			}
		}
		return err
	})

	// NOTE(DEV-2982): This is for debugging the session lost problem
	logger.WithSkipStackTrace().Warn(ctx,
		"delete IDP session",
		slog.String("idp_session_id", session.ID),
		slog.Time("idp_session_created_at", session.CreatedAt),
		slog.String("user_id", session.Attrs.UserID),
		slog.Bool("refresh_token_log", true),
	)

	return
}

func (s *StoreRedis) CleanUpForDeletingUserID(ctx context.Context, userID string) (err error) {
	listKey := sessionListKey(s.AppID, userID)
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, listKey).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return
}

//nolint:gocognit
func (s *StoreRedis) List(ctx context.Context, userID string) (sessions []*IDPSession, err error) {
	logger := StoreRedisLogger.GetLogger(ctx)
	listKey := sessionListKey(s.AppID, userID)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionList, err := conn.HGetAll(ctx, listKey).Result()
		if err != nil {
			return err
		}

		for key := range sessionList {
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
					logger.WithError(err).Error(ctx, "invalid JSON value",
						slog.String("key", key),
					)
					err = nil
				}
			}

			if session == nil {
				// cleanup expired / terminated sessions from the list

				// ignore non-critical error
				_, err = conn.HDel(ctx, listKey, key).Result()
				if err != nil {
					// ignore non-critical error
					logger.WithError(err).Error(ctx, "failed to update session list",
						slog.String("key", listKey),
					)
					err = nil
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
