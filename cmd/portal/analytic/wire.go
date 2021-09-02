//+build wireinject

package analytic

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/google/wire"
)

func NewEmptyAppID() config.AppID {
	return config.AppID("")
}

func NewUserWeeklyReport(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
) *UserWeeklyReport {
	panic(wire.Build(
		// Analytic reports need to run query based on different app ids
		// To simplify the implementation, we assume all apps run with the same
		// database, and the appID will be provided during the runtime
		// so the app id injected through wire will not be used
		NewEmptyAppID,
		DependencySet,
	))
}
