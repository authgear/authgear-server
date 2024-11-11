//go:build wireinject
// +build wireinject

package analytic

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

func NewUserWeeklyReport(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
) *analytic.UserWeeklyReport {
	panic(wire.Build(
		DependencySet,
	))
}

func NewProjectHourlyReport(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
) *analytic.ProjectHourlyReport {
	panic(wire.Build(
		DependencySet,
	))
}

func NewProjectWeeklyReport(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
) *analytic.ProjectWeeklyReport {
	panic(wire.Build(
		DependencySet,
	))
}

func NewProjectMonthlyReport(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
) *analytic.ProjectMonthlyReport {
	panic(wire.Build(
		DependencySet,
	))
}

func NewCountCollector(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
	redisPool *redis.Pool,
	credentials *config.AnalyticRedisCredentials,
) *analytic.CountCollector {
	panic(wire.Build(
		DependencySet,
	))
}

func NewPeriodicalArgumentParser() *periodical.ArgumentParser {
	panic(wire.Build(
		clock.DependencySet,
		periodical.DependencySet,
	))
}

func NewPosthogIntegration(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	auditDatabaseCredentials *config.AuditDatabaseCredentials,
	redisPool *redis.Pool,
	credentials *config.AnalyticRedisCredentials,
	posthogCredentials *analytic.PosthogCredentials,
) *analytic.PosthogIntegration {
	panic(wire.Build(
		DependencySet,
	))
}
