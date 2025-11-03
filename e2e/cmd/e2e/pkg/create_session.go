package e2e

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	infraredis "github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oauthpq "github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (c *End2End) CreateSession(
	ctx context.Context,
	appID string,
	selectUserIDSQL string,
	sessionType session.Type,
	sessionID string,
	clientID string,
	token string) (err error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	dbConn := openDB(cfg.GlobalDatabase.DatabaseURL, cfg.GlobalDatabase.DatabaseSchema)

	vars := map[string]interface{}{
		"AppID": appID,
	}

	tmpl, err := ParseSQLTemplate("sql-select-user-id", selectUserIDSQL)
	if err != nil {
		return fmt.Errorf("failed to parse SQL template: %w", err)
	}

	var parsedsql bytes.Buffer
	if err := tmpl.Execute(&parsedsql, vars); err != nil {
		return fmt.Errorf("failed to execute SQL template: %w", err)
	}

	rows, err := dbConn.Query(parsedsql.String())
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	parsedRows, err := ParseRows(rows)
	rows.Close()
	if err != nil {
		return fmt.Errorf("failed to parse SQL rows: %w", err)
	}
	if len(parsedRows) < 1 {
		return fmt.Errorf("Cannot determine user ID because SQL returned 0 rows.")
	}
	userID, ok := parsedRows[0]["id"].(string)
	if !ok || userID == "" {
		return fmt.Errorf("Cannot determine user ID because SQL returned invalid result.")
	}

	redisPool := infraredis.NewPool()
	redisHub := infraredis.NewHub(ctx, redisPool)
	redis := appredis.NewHandle(
		redisPool,
		redisHub,
		&cfg.RedisConfig,
		&config.RedisCredentials{
			RedisURL: cfg.GlobalRedis.RedisURL,
		},
	)
	clk := clock.NewSystemClock()
	accessEvent := access.NewEvent(
		clk.NowUTC(), "127.0.0.1", "e2e",
	)
	switch sessionType {
	case session.TypeIdentityProvider:
		idpSessionStore := &idpsession.StoreRedis{
			Redis: redis,
			AppID: config.AppID(appID),
			Clock: clk,
		}
		expiry, err := time.Parse(time.RFC3339, "2160-01-02T15:04:05Z") // Not important
		if err != nil {
			return fmt.Errorf("Failed to parse expiry: %w", err)
		}
		encodedToken := idpsession.E2EEncodeToken(sessionID, token)
		tokenHash := idpsession.E2EHashToken(encodedToken)
		err = idpSessionStore.Create(ctx, &idpsession.IDPSession{
			ID:              sessionID,
			AppID:           appID,
			CreatedAt:       clk.NowUTC(),
			AuthenticatedAt: clk.NowUTC(),
			Attrs:           *session.NewAttrs(userID),
			AccessInfo: access.Info{
				InitialAccess: accessEvent,
				LastAccess:    accessEvent,
			},
			TokenHash: tokenHash,
		}, expiry)
		if err != nil {
			return fmt.Errorf("Failed to create idp session: %w", err)
		}
	case session.TypeOfflineGrant:
		encodedToken := oauth.E2EEncodeRefreshToken(sessionID, token)
		tokenHash := oauth.E2EHashToken(encodedToken)
		scopes := []string{oauth.FullAccessScope, oauth.FullUserInfoScope}
		dbCred := &config.DatabaseCredentials{
			DatabaseURL:    cfg.GlobalDatabase.DatabaseURL,
			DatabaseSchema: cfg.GlobalDatabase.DatabaseSchema,
		}
		dbHandle := appdb.NewHandle(db.NewPool(),
			config.NewDefaultDatabaseEnvironmentConfig(),
			dbCred,
		)
		authzStore := oauthpq.AuthorizationStore{
			SQLBuilder:  appdb.NewSQLBuilderApp(dbCred, config.AppID(appID)),
			SQLExecutor: appdb.NewSQLExecutor(dbHandle),
		}
		authz := &oauth.Authorization{
			ID:        uuid.New(),
			AppID:     appID,
			ClientID:  clientID,
			UserID:    userID,
			CreatedAt: clk.NowUTC(),
			UpdatedAt: clk.NowUTC(),
			Scopes:    scopes,
		}
		err := dbHandle.WithTx(ctx, func(ctx context.Context) error {
			return authzStore.Create(ctx, authz)
		})
		if err != nil {
			return fmt.Errorf("Failed to create Authorization: %w", err)
		}
		grant := oauth.OfflineGrant{
			ID:              sessionID,
			AppID:           appID,
			InitialClientID: clientID,
			CreatedAt:       clk.NowUTC(),
			AuthenticatedAt: clk.NowUTC(),
			Attrs:           *session.NewAttrs(userID),
			AccessInfo: access.Info{
				InitialAccess: accessEvent,
				LastAccess:    accessEvent,
			},
			RefreshTokens: []oauth.OfflineGrantRefreshToken{{
				InitialTokenHash: tokenHash,
				ClientID:         clientID,
				CreatedAt:        clk.NowUTC(),
				Scopes:           scopes,
				AuthorizationID:  authz.ID,
				AccessInfo: &access.Info{
					InitialAccess: accessEvent,
					LastAccess:    accessEvent,
				},
			}},
			ExpireAtForResolvedSession: clk.NowUTC().Add(duration.UserInteraction),
		}
		offlineGrantStore := &oauthredis.Store{
			AppID: config.AppID(appID),
			Redis: redis,
			Clock: clk,
		}
		err = offlineGrantStore.CreateOfflineGrant(ctx, &grant)
		if err != nil {
			return fmt.Errorf("Failed to create OfflineGrant: %w", err)
		}
		break
	default:
		return fmt.Errorf("Failed to create: unsupported session type.")
	}

	return nil
}
