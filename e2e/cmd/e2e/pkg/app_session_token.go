package e2e

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const appSessionTokenDuration = 5 * time.Minute

func (c *End2End) GenerateAppSessionToken(ctx context.Context, appID string, refreshToken string) (string, error) {
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

	var appSessionToken string

	err = configSrcController.ResolveContext(ctx, appID, func(ctx context.Context, appCtx *config.AppContext) error {
		_, appProvider := p.NewAppProvider(ctx, appCtx)

		appCfg := appCtx.Config.AppConfig
		secretConfig := appCtx.Config.SecretConfig
		appIDCfg := appCfg.ID
		clk := clock.NewSystemClock()

		// Decode refresh token to get the raw token value and grant ID.
		tokenValue, grantID, err := oauth.DecodeRefreshToken(refreshToken)
		if err != nil {
			return fmt.Errorf("failed to decode refresh token: %w", err)
		}

		// Look up the offline grant directly from Redis.
		redisStore := &oauthredis.Store{
			Redis: appProvider.Redis,
			AppID: appIDCfg,
			Clock: clk,
		}
		offlineGrant, err := redisStore.GetOfflineGrantWithoutExpireAt(ctx, grantID)
		if errors.Is(err, oauth.ErrGrantNotFound) {
			return fmt.Errorf("offline grant not found (id=%s): %w", grantID, err)
		} else if err != nil {
			return fmt.Errorf("failed to get offline grant: %w", err)
		}

		// Verify the token matches the grant.
		tokenHash := oauth.HashToken(tokenValue)
		if !offlineGrant.MatchCurrentHash(tokenHash) {
			return fmt.Errorf("refresh token hash does not match offline grant")
		}

		session, ok := offlineGrant.ToSession(tokenHash)
		if !ok {
			return fmt.Errorf("failed to resolve offline grant session from token hash")
		}

		// Verify the authorization has full-access scope.
		dbCredentials := deps.ProvideDatabaseCredentials(secretConfig)
		sqlBuilderApp := appdb.NewSQLBuilderApp(dbCredentials, appIDCfg)
		sqlExecutor := appdb.NewSQLExecutor(appProvider.AppDatabase)
		authzStore := &oauthpq.AuthorizationStore{
			SQLBuilder:  sqlBuilderApp,
			SQLExecutor: sqlExecutor,
		}

		var authz *oauth.Authorization
		err = appProvider.AppDatabase.ReadOnly(ctx, func(ctx context.Context) error {
			authz, err = authzStore.GetByID(ctx, session.AuthorizationID)
			return err
		})
		if err != nil {
			return fmt.Errorf("failed to get authorization: %w", err)
		}

		if !authz.IsAuthorized([]string{oauth.FullAccessScope}) {
			return fmt.Errorf("authorization does not include full-access scope")
		}

		// Create the app session token in Redis.
		now := clk.NowUTC()
		rawToken := oauth.GenerateToken()
		sToken := &oauth.AppSessionToken{
			AppID:                   string(appIDCfg),
			OfflineGrantID:          offlineGrant.ID,
			CreatedAt:               now,
			ExpireAt:                now.Add(appSessionTokenDuration),
			TokenHash:               oauth.HashToken(rawToken),
			InitialRefreshTokenHash: session.InitialTokenHash,
		}

		err = redisStore.CreateAppSessionToken(ctx, sToken)
		if err != nil {
			return fmt.Errorf("failed to create app session token: %w", err)
		}

		appSessionToken = rawToken
		return nil
	})
	if err != nil {
		return "", err
	}

	return appSessionToken, nil
}
