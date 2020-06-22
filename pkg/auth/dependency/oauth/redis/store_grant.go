package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	"github.com/skygeario/skygear-server/pkg/db"
)

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

type GrantStore struct {
	Context     context.Context
	AppID       string
	Logger      *logrus.Entry
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
	err := s.load(redis.GetConn(s.Context), codeGrantKey(s.AppID, codeHash), g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GrantStore) CreateCodeGrant(grant *oauth.CodeGrant) error {
	return s.save(redis.GetConn(s.Context), codeGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
}

func (s *GrantStore) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	return s.del(redis.GetConn(s.Context), codeGrantKey(grant.AppID, grant.CodeHash))
}

func (s *GrantStore) GetAccessGrant(tokenHash string) (*oauth.AccessGrant, error) {
	g := &oauth.AccessGrant{}
	err := s.load(redis.GetConn(s.Context), accessGrantKey(s.AppID, tokenHash), g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GrantStore) CreateAccessGrant(grant *oauth.AccessGrant) error {
	return s.save(redis.GetConn(s.Context), accessGrantKey(grant.AppID, grant.TokenHash), grant, grant.ExpireAt, true)
}

func (s *GrantStore) DeleteAccessGrant(grant *oauth.AccessGrant) error {
	return s.del(redis.GetConn(s.Context), accessGrantKey(grant.AppID, grant.TokenHash))
}

func (s *GrantStore) GetOfflineGrant(id string) (*oauth.OfflineGrant, error) {
	g := &oauth.OfflineGrant{}
	err := s.load(redis.GetConn(s.Context), offlineGrantKey(s.AppID, id), g)
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

	conn := redis.GetConn(s.Context)

	_, err = conn.Do("HSET", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry)
	if err != nil {
		return fmt.Errorf("failed to update session list: %w", err)
	}

	err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, grant.ExpireAt, true)
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

	conn := redis.GetConn(s.Context)

	_, err = conn.Do("HSET", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry)
	if err != nil {
		return fmt.Errorf("failed to update session list: %w", err)
	}

	err = s.save(conn, offlineGrantKey(grant.AppID, grant.ID), grant, grant.ExpireAt, false)
	if err != nil {
		return err
	}

	return nil
}

func (s *GrantStore) DeleteOfflineGrant(grant *oauth.OfflineGrant) error {
	conn := redis.GetConn(s.Context)

	err := s.del(conn, offlineGrantKey(grant.AppID, grant.ID))
	if err != nil {
		return err
	}

	_, err = conn.Do("HDEL", offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID)
	if err != nil {
		s.Logger.WithError(err).Error("failed to update session list")
	}

	return nil
}

func (s *GrantStore) ListOfflineGrants(userID string) ([]*oauth.OfflineGrant, error) {
	conn := redis.GetConn(s.Context)
	listKey := offlineGrantListKey(s.AppID, userID)

	sessionList, err := redigo.StringMap(conn.Do("HGETALL", listKey))
	if err != nil {
		return nil, err
	}

	var grants []*oauth.OfflineGrant
	for id := range sessionList {
		grant, err := s.GetOfflineGrant(id)

		if errors.Is(err, oauth.ErrGrantNotFound) {
			_, err = conn.Do("HDEL", listKey, id)
			if err != nil {
				s.Logger.WithError(err).Error("failed to update session list")
			}
		} else if err != nil {
			return nil, err
		} else {
			grants = append(grants, grant)
		}
	}

	sort.Slice(grants, func(i, j int) bool {
		return grants[i].ID < grants[j].ID
	})
	return grants, nil
}
