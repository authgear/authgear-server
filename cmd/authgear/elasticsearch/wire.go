//+build wireinject

package elasticsearch

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func NewReindexer(
	ctx context.Context,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
) *Reindexer {
	panic(wire.Build(DependencySet))
}
