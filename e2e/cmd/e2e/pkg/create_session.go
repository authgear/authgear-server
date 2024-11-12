package e2e

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	infraredis "github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func (c *End2End) CreateSession(
	ctx context.Context,
	appID string,
	selectUserIDSQL string,
	sessionType session.Type,
	sessionID string,
	token string) (err error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	db := openDB(cfg.GlobalDatabase.DatabaseURL, cfg.GlobalDatabase.DatabaseSchema)

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

	rows, err := db.Query(parsedsql.String())
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

	lf := log.NewFactory(log.LevelInfo)

	redisPool := infraredis.NewPool()
	redisHub := infraredis.NewHub(ctx, redisPool, lf)
	redis := appredis.NewHandle(
		redisPool,
		redisHub,
		&cfg.RedisConfig,
		&config.RedisCredentials{
			RedisURL: cfg.GlobalRedis.RedisURL,
		},
		lf,
	)
	clk := clock.NewSystemClock()
	switch sessionType {
	case session.TypeIdentityProvider:
		idpSessionStore := &idpsession.StoreRedis{
			Redis:  redis,
			AppID:  config.AppID(appID),
			Clock:  clk,
			Logger: idpsession.NewStoreRedisLogger(lf),
		}
		expiry, err := time.Parse(time.RFC3339, "2160-01-02T15:04:05Z") // Not important
		if err != nil {
			return fmt.Errorf("Failed to parse expiry: %w", err)
		}
		accessEvent := access.NewEvent(
			clk.NowUTC(), "127.0.0.1", "e2e",
		)
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
		// TODO: Implement this when you need to create an offline grant in e2e tests
		fallthrough
	default:
		return fmt.Errorf("Failed to create: unsupported session type.")
	}

	return nil
}
