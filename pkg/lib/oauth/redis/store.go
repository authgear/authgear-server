package redis

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
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var OAuthStoreLogger = slogutil.NewLogger("oauth-store")

type Store struct {
	Redis       *appredis.Handle
	AppID       config.AppID
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

	logger := OAuthStoreLogger.GetLogger(ctx)
	// NOTE(DEV-2982): This is for debugging the session lost problem
	logger.WithSkipLogging().Error(ctx,
		"create offline grant",
		slog.String("offline_grant_id", grant.ID),
		slog.String("offline_grant_initial_client_id", grant.InitialClientID),
		slog.Time("offline_grant_created_at", grant.CreatedAt),
		slog.String("user_id", grant.Attrs.UserID),
		slog.Bool("refresh_token_log", true),
	)

	return nil
}

func (s *Store) UpdateOfflineGrantWithMutator(ctx context.Context, grantID string, expireAt time.Time, mutator func(*oauth.OfflineGrant) *oauth.OfflineGrant) (*oauth.OfflineGrant, error) {
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

	grant = mutator(grant)

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
	options oauth.AddOfflineGrantRefreshTokenOptions,
) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), options.OfflineGrantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, options.OfflineGrantID)
	if err != nil {
		return nil, err
	}

	now := s.Clock.NowUTC()

	newRefreshToken := oauth.OfflineGrantRefreshToken{
		InitialTokenHash: options.TokenHash,
		ClientID:         options.ClientID,
		CreatedAt:        now,
		Scopes:           options.Scopes,
		AuthorizationID:  options.AuthorizationID,
		DPoPJKT:          options.DPoPJKT,
		AccessInfo:       &options.AccessInfo,
		ExpireAt:         options.ShortLivedRefreshTokenExpireAt,
	}

	grant.RefreshTokens = append(grant.RefreshTokens, newRefreshToken)
	err = s.updateOfflineGrant(ctx, grant, options.OfflineGrantExpireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) RotateOfflineGrantRefreshToken(
	ctx context.Context,
	opts oauth.RotateOfflineGrantRefreshTokenOptions,
	expireAt time.Time,
) (*oauth.OfflineGrant, error) {
	mutexName := offlineGrantMutexName(string(s.AppID), opts.OfflineGrantID)
	mutex := s.Redis.NewMutex(mutexName)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = mutex.UnlockContext(ctx)
	}()

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, opts.OfflineGrantID)
	if err != nil {
		return nil, err
	}

	var tokenToRotate *oauth.OfflineGrantRefreshToken
	tokenIndex := -1
	for i, token := range grant.RefreshTokens {
		if token.MatchInitialHash(opts.InitialRefreshTokenHash) {
			tokenToRotate = &grant.RefreshTokens[i]
			tokenIndex = i
			break
		}
	}

	if tokenToRotate == nil {
		return nil, oauth.ErrGrantNotFound
	}

	tokenToRotate.RotatedTokenHash = &opts.NewRefreshTokenHash
	t := s.Clock.NowUTC()
	tokenToRotate.RotatedAt = &t
	grant.RefreshTokens[tokenIndex] = *tokenToRotate

	err = s.updateOfflineGrant(ctx, grant, expireAt)
	if err != nil {
		return nil, err
	}

	return grant, nil
}

func (s *Store) RemoveOfflineGrantRefreshTokens(ctx context.Context, grantID string, initialTokenHashes []string, expireAt time.Time) (*oauth.OfflineGrant, error) {
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
	for _, hash := range initialTokenHashes {
		tokenHashesSet[hash] = hash
	}

	grant, err := s.GetOfflineGrantWithoutExpireAt(ctx, grantID)
	if err != nil {
		return nil, err
	}

	newRefreshTokens := []oauth.OfflineGrantRefreshToken{}
	for _, token := range grant.RefreshTokens {
		token := token
		if _, exist := tokenHashesSet[token.InitialTokenHash]; !exist {
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
	logger := OAuthStoreLogger.GetLogger(ctx)
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		err := s.del(ctx, conn, offlineGrantKey(grant.AppID, grant.ID))
		if err != nil {
			return err
		}

		// NOTE(DEV-2982): This is for debugging the session lost problem
		// TODO(slog): Before we have fine-grained logging, use WithSkipLogging().Error() to force logging to stderr.
		logger.WithSkipLogging().Error(ctx,
			"delete offline grant",
			slog.String("offline_grant_id", grant.ID),
			slog.String("offline_grant_initial_client_id", grant.InitialClientID),
			slog.Time("offline_grant_created_at", grant.CreatedAt),
			slog.String("user_id", grant.Attrs.UserID),
			slog.Bool("refresh_token_log", true),
		)

		_, err = conn.HDel(ctx, offlineGrantListKey(grant.AppID, grant.Attrs.UserID), grant.ID).Result()
		if err != nil {
			// Ignore err
			logger.WithError(err).Error(ctx, "failed to update session list")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListOfflineGrants(ctx context.Context, userID string) ([]*oauth.OfflineGrant, error) {
	logger := OAuthStoreLogger.GetLogger(ctx)
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
					// ignore non-critical error
					logger.WithError(err).Error(ctx, "failed to update session list")
					continue
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
