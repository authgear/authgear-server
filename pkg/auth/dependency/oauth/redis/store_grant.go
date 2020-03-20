package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/redis"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

type GrantStore struct {
	Context     context.Context
	AppID       string
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
	Time        coretime.Provider
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
	ttl := expireAt.Sub(s.Time.NowUTC())

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
