//+build wireinject

package elasticsearch

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func NewQuery(
	ctx context.Context,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
) *Query {
	panic(wire.Build(DependencySet))
}
