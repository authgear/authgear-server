package e2e

import (
	"context"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const refreshTokenLifetime = 30 * 24 * time.Hour

func (c *End2End) GenerateRefreshToken(ctx context.Context, appID, userID, clientID string) (string, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return "", err
	}
	cfg.ConfigSource = &configsource.Config{
		Type:  configsource.TypeDatabase,
		Watch: false,
	}

	ctx, p, err := deps.NewRootProvider(
		ctx,
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		return "", err
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		return "", err
	}
	defer configSrcController.Close()

	var encodedToken string

	err = configSrcController.ResolveContext(ctx, appID, func(ctx context.Context, appCtx *config.AppContext) error {
		_, appProvider := p.NewAppProvider(ctx, appCtx)

		appCfg := appCtx.Config.AppConfig
		secretConfig := appCtx.Config.SecretConfig
		appIDCfg := appCfg.ID
		clk := clock.NewSystemClock()
		now := clk.NowUTC()

		dbCredentials := deps.ProvideDatabaseCredentials(secretConfig)
		sqlBuilderApp := appdb.NewSQLBuilderApp(dbCredentials, appIDCfg)
		sqlExecutor := appdb.NewSQLExecutor(appProvider.AppDatabase)
		authzStore := &oauthpq.AuthorizationStore{
			SQLBuilder:  sqlBuilderApp,
			SQLExecutor: sqlExecutor,
		}

		scopes := []string{oauth.ScopeOpenID, oauth.OfflineAccess, oauth.FullAccessScope}

		var authz *oauth.Authorization
		err := appProvider.AppDatabase.WithTx(ctx, func(ctx context.Context) error {
			existing, err := authzStore.Get(ctx, userID, clientID)
			if err != nil && !errors.Is(err, oauth.ErrAuthorizationNotFound) {
				return err
			}
			if existing != nil {
				authz = existing
				return nil
			}
			authz = &oauth.Authorization{
				ID:        uuid.New(),
				AppID:     string(appIDCfg),
				ClientID:  clientID,
				UserID:    userID,
				CreatedAt: now,
				UpdatedAt: now,
				Scopes:    scopes,
			}
			return authzStore.Create(ctx, authz)
		})
		if err != nil {
			return err
		}

		rawToken := oauth.GenerateToken()
		tokenHash := oauth.HashToken(rawToken)

		refreshToken := oauth.OfflineGrantRefreshToken{
			InitialTokenHash: tokenHash,
			ClientID:         clientID,
			CreatedAt:        now,
			Scopes:           scopes,
			AuthorizationID:  authz.ID,
			AccessInfo:       &access.Info{},
		}

		accessEvent := access.NewEvent(now, httputil.RemoteIP(""), httputil.UserAgentString(""))
		grantID := uuid.New()
		offlineGrant := &oauth.OfflineGrant{
			AppID:           string(appIDCfg),
			ID:              grantID,
			InitialClientID: clientID,
			CreatedAt:       now,
			AuthenticatedAt: now,
			Attrs: session.Attrs{
				UserID: userID,
				Claims: map[model.ClaimName]any{},
			},
			AccessInfo: access.Info{
				InitialAccess: accessEvent,
				LastAccess:    accessEvent,
			},
			RefreshTokens: []oauth.OfflineGrantRefreshToken{refreshToken},

			// Deprecated fields kept for backward compatibility with MatchCurrentHash.
			Deprecated_AuthorizationID: authz.ID,
			Deprecated_Scopes:          scopes,
			Deprecated_TokenHash:       tokenHash,

			ExpireAtForResolvedSession: now.Add(refreshTokenLifetime),
		}

		redisStore := &oauthredis.Store{
			Redis: appProvider.Redis,
			AppID: appIDCfg,
			Clock: clk,
		}

		if err := redisStore.CreateOfflineGrant(ctx, offlineGrant); err != nil {
			return err
		}

		encodedToken = oauth.EncodeRefreshToken(rawToken, grantID)
		return nil
	})
	if err != nil {
		return "", err
	}

	return encodedToken, nil
}
