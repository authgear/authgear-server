//go:build wireinject
// +build wireinject

package usage

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/google/wire"
)

func NewCountCollector(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	redisPool *redis.Pool,
	credentials *config.AnalyticRedisCredentials,
) *usage.CountCollector {
	panic(wire.Build(
		DependencySet,
	))
}
