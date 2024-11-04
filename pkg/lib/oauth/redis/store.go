package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
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
	Redis       *appredis.Handle
	AppID       config.AppID
	Logger      Logger
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
}

func (s *Store) loadData(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string) ([]byte, error) {
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

func (s *Store) save(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, value interface{}, expireAt time.Time, ifNotExists bool) error {
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

func (s *Store) del(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string) error {
	_, err := conn.Del(ctx, key).Result()
	return err
}

func (s *Store) GetCodeGrant(ctx context.Context, codeHash string) (*oauth.CodeGrant, error) {
	var g *oauth.CodeGrant
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, codeGrantKey(string(s.AppID), codeHash))
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

func (s *Store) CreateCodeGrant(ctx context.Context, grant *oauth.CodeGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, codeGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteCodeGrant(ctx context.Context, grant *oauth.CodeGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.del(ctx, conn, codeGrantKey(grant.AppID, grant.CodeHash))
	})
}

func (s *Store) GetSettingsActionGrant(ctx context.Context, codeHash string) (*oauth.SettingsActionGrant, error) {
	var g *oauth.SettingsActionGrant
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, settingsActionGrantKey(string(s.AppID), codeHash))
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

func (s *Store) CreateSettingsActionGrant(ctx context.Context, grant *oauth.SettingsActionGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, settingsActionGrantKey(grant.AppID, grant.CodeHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteSettingsActionGrant(ctx context.Context, grant *oauth.SettingsActionGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.del(ctx, conn, settingsActionGrantKey(grant.AppID, grant.CodeHash))
	})
}

func (s *Store) GetAccessGrant(ctx context.Context, tokenHash string) (*oauth.AccessGrant, error) {
	var g *oauth.AccessGrant
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, accessGrantKey(string(s.AppID), tokenHash))
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

func (s *Store) CreateAccessGrant(ctx context.Context, grant *oauth.AccessGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, accessGrantKey(grant.AppID, grant.TokenHash), grant, grant.ExpireAt, true)
	})
}

func (s *Store) DeleteAccessGrant(ctx context.Context, grant *oauth.AccessGrant) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.del(ctx, conn, accessGrantKey(grant.AppID, grant.TokenHash))
	})
}

func (s *Store) GetOfflineGrantWithoutExpireAt(ctx context.Context, id string) (*oauth.OfflineGrant, error) {
	var g *oauth.OfflineGrant
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, offlineGrantKey(string(s.AppID), id))
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

func (s *Store) CreateOfflineGrant(ctx context.Context, grant *oauth.OfflineGrant) error {
	expiry, err := grant.ExpireAtForResolvedSession.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err = conn.HSet(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(ctx, conn, offlineGrantKey(grant.AppID, grant.ID), grant, grant.ExpireAtForResolvedSession, true)
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

// UpdateOfflineGrantLastAccess updates the last access event for an offline grant
func (s *Store) UpdateOfflineGrantLastAccess(ctx context.Context, grantID string, accessEvent access.Event, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	grant.AccessInfo.LastAccess = accessEvent

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantDeviceInfo(ctx context.Context, grantID string, deviceInfo map[string]interface{}, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	grant.DeviceInfo = deviceInfo

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantAuthenticatedAt(ctx context.Context, grantID string, authenticatedAt time.Time, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	grant.AuthenticatedAt = authenticatedAt

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantApp2AppDeviceKey(ctx context.Context, grantID string, newKey string, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	grant.App2AppDeviceKeyJWKJSON = newKey

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) UpdateOfflineGrantDeviceSecretHash(
	ctx context.Context,
	grantID string,
	newDeviceSecretHash string,
	dpopJKT string,
	expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	grant.DeviceSecretHash = newDeviceSecretHash
	grant.DeviceSecretDPoPJKT = dpopJKT

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) AddOfflineGrantSAMLServiceProviderParticipant(
	ctx context.Context,
	grantID string,
	newServiceProviderID string,
	expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	newParticipatedSAMLServiceProviderIDs := grant.GetParticipatedSAMLServiceProviderIDsSet()
	newParticipatedSAMLServiceProviderIDs.Add(newServiceProviderID)
	grant.ParticipatedSAMLServiceProviderIDs = newParticipatedSAMLServiceProviderIDs.Keys()
	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) AddOfflineGrantRefreshToken(
	ctx context.Context,
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
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
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
	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) RemoveOfflineGrantRefreshTokens(ctx context.Context, grantID string, tokenHashes []string, expireAt time.Time) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), grantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	tokenHashesSet := map[string]interface{}{}
	for _, hash := range tokenHashes {
		tokenHashesSet[hash] = hash
	}

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
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
		err = s.updateOfflineGrant(ctx, grant, expireAt)
		if err != nil {
			return nil, err
		}
	} else {
		// Remove the offline grant if it has no valid tokens
		err = s.DeleteOfflineGrant(ctx, grant)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return grant, nil
}

func (s *Store) updateOfflineGrant(ctx context.Context, grant *oauth.OfflineGrant, expireAt time.Time) error {
	grant.ExpireAtForResolvedSession = expireAt
	expiry, err := expireAt.MarshalText()
	if err != nil {
		return err
	}

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err = conn.HSet(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID, expiry).Result()
		if err != nil {
			return fmt.Errorf("failed to update session list: %w", err)
		}

		err = s.save(ctx, conn, offlineGrantKey(grant.AppID, grant.ID), grant, expireAt, false)
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

func (s *Store) DeleteOfflineGrant(ctx context.Context, grant *oauth.OfflineGrant) error {
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		err := s.del(ctx, conn, offlineGrantKey(grant.AppID, grant.ID))
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

func (s *Store) ListOfflineGrants(ctx context.Context, userID string) ([]*oauth.OfflineGrant, error) {
	listKey := offlineGrantListKey(string(s.AppID), userID)

	var grants []*oauth.OfflineGrant
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionList, err := conn.HGetAll(ctx, listKey).Result()
		if err != nil {
			return err
		}

		for id := range sessionList {
			var data []byte
			data, err = s.loadData(ctx, conn, offlineGrantKey(string(s.AppID), id))
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

func (s *Store) ListClientOfflineGrants(ctx context.Context, clientID string, userID string) ([]*oauth.OfflineGrant, error) {
	offlineGrants, err := s.ListOfflineGrants(ctx, userID)
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

func (s *Store) GetAppSessionToken(ctx context.Context, tokenHash string) (*oauth.AppSessionToken, error) {
	t := &oauth.AppSessionToken{}

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, appSessionTokenKey(string(s.AppID), tokenHash))
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

func (s *Store) CreateAppSessionToken(ctx context.Context, token *oauth.AppSessionToken) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, appSessionTokenKey(token.AppID, token.TokenHash), token, token.ExpireAt, true)
	})
}

func (s *Store) DeleteAppSessionToken(ctx context.Context, token *oauth.AppSessionToken) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.del(ctx, conn, appSessionTokenKey(token.AppID, token.TokenHash))
	})
}

func (s *Store) GetAppSession(ctx context.Context, tokenHash string) (*oauth.AppSession, error) {
	var t *oauth.AppSession

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := s.loadData(ctx, conn, appSessionKey(string(s.AppID), tokenHash))
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

func (s *Store) CreateAppSession(ctx context.Context, session *oauth.AppSession) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, appSessionKey(session.AppID, session.TokenHash), session, session.ExpireAt, true)
	})
}

func (s *Store) DeleteAppSession(ctx context.Context, session *oauth.AppSession) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.del(ctx, conn, appSessionKey(session.AppID, session.TokenHash))
	})
}

func (s *Store) CreatePreAuthenticatedURLToken(ctx context.Context, token *oauth.PreAuthenticatedURLToken) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return s.save(ctx, conn, preAuthenticatedURLTokenKey(token.AppID, token.TokenHash), token, token.ExpireAt, true)
	})
}

func (s *Store) ConsumePreAuthenticatedURLToken(ctx context.Context, tokenHash string) (*oauth.PreAuthenticatedURLToken, error) {
	t := &oauth.PreAuthenticatedURLToken{}

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := preAuthenticatedURLTokenKey(string(s.AppID), tokenHash)
		data, err := s.loadData(ctx, conn, key)
		if err != nil {
			return err
		}

		t, err = s.unmarshalPreAuthenticatedURLToken(data)
		if err != nil {
			return err
		}

		return s.del(ctx, conn, key)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Store) CleanUpForDeletingUserID(ctx context.Context, userID string) (err error) {
	listKey := offlineGrantListKey(string(s.AppID), userID)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, listKey).Result()
		if err != nil {
			return err
		}
		return nil
	})
	return
}
