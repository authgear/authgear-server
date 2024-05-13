//go:build wireinject
// +build wireinject

package usage

import (
	"context"

	"github.com/getsentry/sentry-go"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/usage"
)

func NewCountCollector(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
	redisPool *redis.Pool,
	credentials *config.AnalyticRedisCredentials,
	hub *sentry.Hub,
) *usage.CountCollector {
	panic(wire.Build(
		DependencySet,
	))
}
