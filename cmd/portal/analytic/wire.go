//+build wireinject

package analytic

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/google/wire"
)

func NewUserWeeklyReport(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
) *UserWeeklyReport {
	panic(wire.Build(DependencySet))
}
