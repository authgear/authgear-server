//+build wireinject

package elasticsearch

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func NewAppLister(
	ctx context.Context,
	databaseCredentials *config.DatabaseCredentials,
) *AppLister {
	panic(wire.Build(DependencySet))
}

func NewReindexer(
	ctx context.Context,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
) *Reindexer {
	panic(wire.Build(DependencySet))
}
