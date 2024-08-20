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
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("oauth-store")}
}

type Store struct {
	Context     context.Context
	Redis       *appredis.Handle
	AppID       config.AppID
	Logger      Logger
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
}

func (s *Store) loadData(conn *goredis.Conn, key string) ([]byte, error) {
	ctx := context.Background()
	data, err := conn.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, oauth.ErrGrantNotFound
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) unmarshalCodeGrant(data []byte) (*oauth.CodeGrant, error) {
	var g oauth.CodeGrant
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) unmarshalSettingsActionGrant(data []byte) (*oauth.SettingsActionGrant, error) {
	var g oauth.SettingsActionGrant
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) unmarshalAccessGrant(data []byte) (*oauth.AccessGrant, error) {
	var g oauth.AccessGrant
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) unmarshalAppSession(data []byte) (*oauth.AppSession, error) {
	var t oauth.AppSession
	err := json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) unmarshalOfflineGrant(data []byte) (*oauth.OfflineGrant, error) {
	var g oauth.OfflineGrant
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}
	if g.AuthenticatedAt.IsZero() {
		g.AuthenticatedAt = g.CreatedAt
	}
	return &g, nil
}

func (s *Store) unmarshalAppSessionToken(data []byte) (*oauth.AppSessionToken, error) {
	var t oauth.AppSessionToken
	err := json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) unmarshalPreAuthenticatedURLToken(data []byte) (*oauth.PreAuthenticatedURLToken, error) {
	var t oauth.PreAuthenticatedURLToken
	err := json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
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
	var g *oauth.CodeGrant
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, codeGrantKey(string(s.AppID), codeHash))
		if err != nil {
			return err
		}
		g, err = s.unmarshalCodeGrant(data)
		if err != nil {
			return err
		}
		return nil
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

func (s *Store) GetSettingsActionGrant(codeHash string) (*oauth.SettingsActionGrant, error) {
	var g *oauth.SettingsActionGrant
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, settingsActionGrantKey(string(s.AppID), codeHash))
		if err != nil {
			return err
		}
		g, err = s.unmarshalSettingsActionGrant(data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *Store) CreateSettingsActionGrant(grant *oauth.SettingsActionGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, settingsActionGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteSettingsActionGrant(grant *oauth.SettingsActionGrant) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.del(conn, settingsActionGrantKey(grant.AppID, grant.CodeHash))
	})
}

func (s *Store) GetAccessGrant(tokenHash string) (*oauth.AccessGrant, error) {
	var g *oauth.AccessGrant
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, accessGrantKey(string(s.AppID), tokenHash))
		if err != nil {
			return err
		}

		g, err = s.unmarshalAccessGrant(data)
		if err != nil {
			return err
		}

		return nil
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

func (s *Store) GetOfflineGrantWithoutExpireAt(id string) (*oauth.OfflineGrant, error) {
	var g *oauth.OfflineGrant
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, offlineGrantKey(string(s.AppID), id))
		if err != nil {
			return err
		}

		g, err = s.unmarshalOfflineGrant(data)
		if err != nil {
			return err
		}

		return nil
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

func (s *Store) AccessWithID(grantID string, accessEvent access.Event, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	grant.AccessInfo.LastAccess = accessEvent

	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) AccessOfflineGrantAndUpdateDeviceInfo(grantID string, accessEvent access.Event, deviceInfo map[string]interface{}, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	grant.AccessInfo.LastAccess = accessEvent
	grant.DeviceInfo = deviceInfo

	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantAuthenticatedAt(grantID string, authenticatedAt time.Time, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	grant.AuthenticatedAt = authenticatedAt

	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantApp2AppDeviceKey(grantID string, newKey string, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	grant.App2AppDeviceKeyJWKJSON = newKey

	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantDeviceSecretHash(
	grantID string,
	newDeviceSecretHash string,
	dpopJKT string,
	expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	grant.DeviceSecretHash = newDeviceSecretHash
	grant.DeviceSecretDPoPJKT = dpopJKT

	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) AddOfflineGrantRefreshToken(
	grantID string,
	expireAt time.Time,
	tokenHash string,
	clientID string,
	scopes []string,
	authorizationID string,
	dpopJKT string,
) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	now := s.Clock.NowUTC()
	newRefreshToken := oauth.OfflineGrantRefreshToken{
		TokenHash:       tokenHash,
		ClientID:        clientID,
		CreatedAt:       now,
		Scopes:          scopes,
		AuthorizationID: authorizationID,
		DPoPJKT:         dpopJKT,
	}

	grant.RefreshTokens = append(grant.RefreshTokens, newRefreshToken)
	err = s.updateOfflineGrant(grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) RemoveOfflineGrantRefreshTokens(grantID string, tokenHashes []string, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(s.Context)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(s.Context)
	}()

	tokenHashesSet := map[string]interface{}{}
	for _, hash := range tokenHashes {
		tokenHashesSet[hash] = hash
	}

	grant, err := s.GetOfflineGrantWithoutExpireAt(grantID)
	if err != nil {
		return nil, err
	}

	newRefreshTokens := []oauth.OfflineGrantRefreshToken{}
	for _, token := range grant.RefreshTokens {
		token := token
		if _, exist := tokenHashesSet[token.TokenHash]; !exist {
			newRefreshTokens = append(newRefreshTokens, token)
		}
	}

	grant.RefreshTokens = newRefreshTokens
	if grant.HasValidTokens() {
		err = s.updateOfflineGrant(grant, expireAt)
		if err != nil {
			return nil, err
		}
	} else {
		// Remove the offline grant if it has no valid tokens
		err = s.DeleteOfflineGrant(grant)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return grant, nil
}

func (s *Store) updateOfflineGrant(grant *oauth.OfflineGrant, expireAt time.Time) error {
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
			var data []byte
			data, err = s.loadData(conn, offlineGrantKey(string(s.AppID), id))
			if errors.Is(err, oauth.ErrGrantNotFound) {
				_, err = conn.HDel(ctx, listKey, id).Result()
				if err != nil {
					s.Logger.WithError(err).Error("failed to update session list")
				}
			} else if err != nil {
				return err
			} else {
				var grant *oauth.OfflineGrant
				grant, err = s.unmarshalOfflineGrant(data)
				if err != nil {
					return err
				}
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

func (s *Store) ListClientOfflineGrants(clientID string, userID string) ([]*oauth.OfflineGrant, error) {
	offlineGrants, err := s.ListOfflineGrants(userID)
	if err != nil {
		return nil, err
	}
	result := []*oauth.OfflineGrant{}
	for _, offlineGrant := range offlineGrants {
		if offlineGrant.HasClientID(clientID) {
			result = append(result, offlineGrant)
		} else {
			for _, token := range offlineGrant.RefreshTokens {
				if token.ClientID == clientID {
					result = append(result, offlineGrant)
					break
				}
			}
		}
	}
	return result, nil
}

func (s *Store) GetAppSessionToken(tokenHash string) (*oauth.AppSessionToken, error) {
	t := &oauth.AppSessionToken{}

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, appSessionTokenKey(string(s.AppID), tokenHash))
		if err != nil {
			return err
		}

		t, err = s.unmarshalAppSessionToken(data)
		if err != nil {
			return err
		}

		return nil
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
	var t *oauth.AppSession

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := s.loadData(conn, appSessionKey(string(s.AppID), tokenHash))
		if err != nil {
			return err
		}

		t, err = s.unmarshalAppSession(data)
		if err != nil {
			return err
		}

		return nil
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

func (s *Store) CreatePreAuthenticatedURLToken(token *oauth.PreAuthenticatedURLToken) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		return s.save(conn, preAuthenticatedURLTokenKey(token.AppID, token.TokenHash), token, token.ExpireAt, true)
	})
}

func (s *Store) ConsumePreAuthenticatedURLToken(tokenHash string) (*oauth.PreAuthenticatedURLToken, error) {
	t := &oauth.PreAuthenticatedURLToken{}

	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		key := preAuthenticatedURLTokenKey(string(s.AppID), tokenHash)
		data, err := s.loadData(conn, key)
		if err != nil {
			return err
		}

		t, err = s.unmarshalPreAuthenticatedURLToken(data)
		if err != nil {
			return err
		}

		return s.del(conn, key)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Store) CleanUpForDeletingUserID(userID string) (err error) {
	ctx := context.Background()
	listKey := offlineGrantListKey(string(s.AppID), userID)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.Del(ctx, listKey).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return
}
