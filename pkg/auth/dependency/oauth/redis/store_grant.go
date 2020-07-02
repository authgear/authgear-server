package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/redis"
)

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("oauth-grant-store")}
}

type GrantStore struct {
	Redis       *redis.Handle
	AppID       config.AppID
	Logger      Logger
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
	Clock       clock.Clock
}

func (s *GrantStore) load(conn redigo.Conn, key string, ptr interface{}) error {
	data, err := redigo.Bytes(conn.Do("GET", key))
	if errors.Is(err, redigo.ErrNil) {
		return oauth.ErrGrantNotFound
	} else if err != nil {
		return err
	}
	return json.Unmarshal(data, &ptr)
}

func (s *GrantStore) save(conn redigo.Conn, key string, value interface{}, expireAt time.Time, ifNotExists bool) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ttl := expireAt.Sub(s.Clock.NowUTC())

	var ctrl string
	if ifNotExists {
		ctrl = "NX"
	} else {
		ctrl = "XX"
	}

	_, err = redigo.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), ctrl))
	if errors.Is(err, redigo.ErrNil) {
		if ifNotExists {
			return errors.New("grant already exist")
		}
		return oauth.ErrGrantNotFound
	}

	return nil
}

func (s *GrantStore) del(conn redigo.Conn, key string) error {
	_, err := conn.Do("DEL", key)
	return err
}

func (s *GrantStore) GetCodeGrant(codeHash string) (*oauth.CodeGrant, error) {
	g := &oauth.CodeGrant{}

	err := s.Redis.WithConn(func(conn redis.Conn) error {
		return s.load(conn, codeGrantKey(string(s.AppID), codeHash), g)
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *GrantStore) CreateCodeGrant(grant *oauth.CodeGrant) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return s.save(conn, codeGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
	})
}

func (s *GrantStore) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return s.del(conn, codeGrantKey(grant.AppID, grant.CodeHash))
	})
}

func (s *GrantStore) GetAccessGrant(tokenHash string) (*oauth.AccessGrant, error) {
	g := &oauth.AccessGrant{}
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		return s.load(conn, accessGrantKey(string(s.AppID), tokenHash), g)
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *GrantStore) CreateAccessGrant(grant *oauth.AccessGrant) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return s.save(conn, accessGrantKey(grant.AppID, grant.TokenHash), grant, grant.ExpireAt, true)
	})
}

func (s *GrantStore) DeleteAccessGrant(grant *oauth.AccessGrant) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		return s.del(conn, accessGrantKey(grant.AppID, grant.TokenHash))
	})
}

func (s *GrantStore) GetOfflineGrant(id string) (*oauth.OfflineGrant, error) {
	g := &oauth.OfflineGrant{}
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		return s.load(conn, offlineGrantKey(string(s.AppID), id), g)
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GrantStore) CreateOfflineGrant(grant *oauth.OfflineGrant) error {
	expiry, err := grant.ExpireAt.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = conn.Do("HSET", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry)
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, grant.ExpireAt, true)
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

func (s *GrantStore) UpdateOfflineGrant(grant *oauth.OfflineGrant) error {
	expiry, err := grant.ExpireAt.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = conn.Do("HSET", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry)
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, grant.ExpireAt, false)
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

func (s *GrantStore) DeleteOfflineGrant(grant *oauth.OfflineGrant) error {
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		err := s.del(conn, offlineGrantKey(grant.AppID, grant.ID))
		if err != nil {
			return err
		}
		_, err = conn.Do("HDEL", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID)
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

func (s *GrantStore) ListOfflineGrants(userID string) ([]*oauth.OfflineGrant, error) {
	listKey := offlineGrantListKey(string(s.AppID), userID)

	var grants []*oauth.OfflineGrant
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		sessionList, err := redigo.StringMap(conn.Do("HGETALL", listKey))
		if err != nil {
			return err
		}

		for id := range sessionList {
			grant := &oauth.OfflineGrant{}
			err = s.load(conn, offlineGrantKey(string(s.AppID), id), grant)
			if errors.Is(err, oauth.ErrGrantNotFound) {
				_, err = conn.Do("HDEL", listKey, id)
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
