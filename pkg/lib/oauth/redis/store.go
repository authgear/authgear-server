package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("oauth-store")}
}

type Store struct {
	Redis       *redis.Handle
	AppID       config.AppID
	Logger      Logger
	SQLBuilder  db.SQLBuilder
	SQLExecutor *tenantdb.SQLExecutor
	Clock       clock.Clock
}

func (s *Store) load(conn *goredis.Conn, key string, ptr interface{}) error {
	ctx := context.Background()
	data, err := conn.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return oauth.ErrGrantNotFound
	} else if err != nil {
		return err
	}
	return json.Unmarshal(data, &ptr)
}

func (s *Store) save(conn *goredis.Conn, key string, value interface{}, expireAt time.Time, ifNotExists bool) error {
	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttl := expireAt.Sub(s.Clock.NowUTC())

	if ifNotExists {
		_, err = conn.SetNX(ctx, key, data, ttl).Result()
	} else {
		_, err = conn.SetXX(ctx, key, data, ttl).Result()
	}
	if errors.Is(err, goredis.Nil) {
		if ifNotExists {
			return errors.New("grant already exist")
		}
		return oauth.ErrGrantNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Store) del(conn *goredis.Conn, key string) error {
	ctx := context.Background()
	_, err := conn.Del(ctx, key).Result()
	return err
}

func (s *Store) GetCodeGrant(codeHash string) (*oauth.CodeGrant, error) {
	g := &oauth.CodeGrant{}

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.load(conn, codeGrantKey(string(s.AppID), codeHash), g)
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *Store) CreateCodeGrant(grant *oauth.CodeGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, codeGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.del(conn, codeGrantKey(grant.AppID, grant.CodeHash))
	})
}

func (s *Store) GetAccessGrant(tokenHash string) (*oauth.AccessGrant, error) {
	g := &oauth.AccessGrant{}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.load(conn, accessGrantKey(string(s.AppID), tokenHash), g)
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *Store) CreateAccessGrant(grant *oauth.AccessGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, accessGrantKey(grant.AppID, grant.TokenHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteAccessGrant(grant *oauth.AccessGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.del(conn, accessGrantKey(grant.AppID, grant.TokenHash))
	})
}

func (s *Store) GetOfflineGrant(id string) (*oauth.OfflineGrant, error) {
	g := &oauth.OfflineGrant{}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.load(conn, offlineGrantKey(string(s.AppID), id), g)
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Store) CreateOfflineGrant(grant *oauth.OfflineGrant, expireAt time.Time) error {
	ctx := context.Background()
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err = conn.HSet(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, expireAt, true)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateOfflineGrant(grant *oauth.OfflineGrant, expireAt time.Time) error {
	ctx := context.Background()
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err = conn.HSet(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, expireAt, false)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteOfflineGrant(grant *oauth.OfflineGrant) error {
	ctx := context.Background()
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		err := s.del(conn, offlineGrantKey(grant.AppID, grant.ID))
		if err != nil {
			return err
		}
		_, err = conn.HDel(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID).Result()
		if err != nil {
			// Ignore err
			s.Logger.WithError(err).Error("failed to update session list")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListOfflineGrants(userID string) ([]*oauth.OfflineGrant, error) {
	ctx := context.Background()
	listKey := offlineGrantListKey(string(s.AppID), userID)

	var grants []*oauth.OfflineGrant
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		sessionList, err := conn.HGetAll(ctx, listKey).Result()
		if err != nil {
			return err
		}

		for id := range sessionList {
			grant := &oauth.OfflineGrant{}
			err = s.load(conn, offlineGrantKey(string(s.AppID), id), grant)
			if errors.Is(err, oauth.ErrGrantNotFound) {
				_, err = conn.HDel(ctx, listKey, id).Result()
				if err != nil {
					s.Logger.WithError(err).Error("failed to update session list")
				}
			} else if err != nil {
				return err
			} else {
				grants = append(grants, grant)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(grants, func(i, j int) bool {
		return grants[i].ID < grants[j].ID
	})
	return grants, nil
}

func (s *Store) GetAppSessionToken(tokenHash string) (*oauth.AppSessionToken, error) {
	t := &oauth.AppSessionToken{}

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.load(conn, appSessionTokenKey(string(s.AppID), tokenHash), t)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Store) CreateAppSessionToken(token *oauth.AppSessionToken) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, appSessionTokenKey(token.AppID, token.TokenHash), token, token.ExpireAt, true)
	})
}

func (s *Store) DeleteAppSessionToken(token *oauth.AppSessionToken) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.del(conn, appSessionTokenKey(token.AppID, token.TokenHash))
	})
}

func (s *Store) GetAppSession(tokenHash string) (*oauth.AppSession, error) {
	t := &oauth.AppSession{}

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.load(conn, appSessionKey(string(s.AppID), tokenHash), t)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Store) CreateAppSession(session *oauth.AppSession) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, appSessionKey(session.AppID, session.TokenHash), session, session.ExpireAt, true)
	})
}

func (s *Store) DeleteAppSession(session *oauth.AppSession) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.del(conn, appSessionKey(session.AppID, session.TokenHash))
	})
}
