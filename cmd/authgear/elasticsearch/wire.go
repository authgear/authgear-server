//+build wireinject

package elasticsearch

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
)

func NewAppLister(
	ctx context.Context,
	databaseCredentials *config.DatabaseCredentials,
) *AppLister {
	panic(wire.Build(DependencySet))
}

func NewReindexer(
	ctx context.Context,
	pool *tenantdb.Pool,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
) *Reindexer {
	panic(wire.Build(DependencySet))
}
