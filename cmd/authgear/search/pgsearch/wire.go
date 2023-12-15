//go:build wireinject
// +build wireinject

package pgsearch

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func NewReindexer(
	ctx context.Context,
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
	appID config.AppID,
) *Reindexer {
	panic(wire.Build(DependencySet))
}
