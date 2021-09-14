//+build wireinject

package analytic

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/google/wire"
)

func NewEmptyAppID() config.AppID {
	// Analytic reports need to run query based on different app ids
	// To simplify the implementation, we assume all apps run with the same
	// database, and the appID will be provided during the runtime
	// so the app id injected through wire will not be used
	return config.AppID("")
}

func NewUserWeeklyReport(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
) *analytic.UserWeeklyReport {
	panic(wire.Build(
		NewEmptyAppID,
		DependencySet,
	))
}

func NewProjectWeeklyReport(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
) *analytic.ProjectWeeklyReport {
	panic(wire.Build(
		NewEmptyAppID,
		DependencySet,
	))
}

func NewCountCollector(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
	redisPool *redis.Pool,
	credentials *config.AnalyticRedisCredentials,
) *analytic.CountCollector {
	panic(wire.Build(
		NewEmptyAppID,
		DependencySet,
	))
}
