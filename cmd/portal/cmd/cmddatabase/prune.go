package cmddatabase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var pruneConfigSourceLogger = slogutil.NewLogger("prune-config-source")

// pruneConfigSource scrubs each app's _portal_config_source row down to the
// minimum authgear.yaml/authgear.secrets.yaml needed to keep the app id
// from ever being re-registered (see configsource.TrimDataForPrune), rather
// than deleting the row like the rest of prune's tables.
func pruneConfigSource(ctx context.Context, dbURL string, dbSchema string, appIDs []string, dryRun bool) error {
	pool := db.NewPool()
	handle := &globaldb.Handle{
		HookHandle: db.NewHookHandle(
			pool,
			db.ConnectionInfo{
				Purpose:     db.ConnectionPurposeGlobal,
				DatabaseURL: dbURL,
			},
			db.ConnectionOptions{
				MaxOpenConnection:     1,
				MaxIdleConnection:     1,
				MaxConnectionLifetime: 1800 * time.Second,
				IdleConnectionTimeout: 300 * time.Second,
			},
		),
	}
	store := &configsource.Store{
		SQLBuilder:  &globaldb.SQLBuilder{SQLBuilder: db.NewSQLBuilder(dbSchema)},
		SQLExecutor: &globaldb.SQLExecutor{},
	}

	logger := pruneConfigSourceLogger.GetLogger(ctx)

	prune := func(ctx context.Context) error {
		for _, appID := range appIDs {
			dbs, err := store.GetDatabaseSourceByAppID(ctx, appID)
			if errors.Is(err, configsource.ErrAppNotFound) {
				continue
			} else if err != nil {
				return err
			}

			trimmed, err := configsource.TrimDataForPrune(ctx, dbs.Data)
			if err != nil {
				return err
			}

			if dryRun {
				logger.Info(ctx, "would trim config source data", slog.String("app_id", appID))
				continue
			}

			dbs.Data = trimmed
			dbs.UpdatedAt = time.Now().UTC()
			if err := store.UpdateDatabaseSource(ctx, dbs); err != nil {
				return err
			}
			logger.Info(ctx, "trimmed config source data", slog.String("app_id", appID))
		}
		return nil
	}

	if dryRun {
		return handle.ReadOnly(ctx, prune)
	}
	return handle.WithTx(ctx, prune)
}
