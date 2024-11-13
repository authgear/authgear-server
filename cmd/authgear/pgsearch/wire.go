//go:build wireinject
// +build wireinject

package pgsearch

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func NewReindexer(
	pool *db.Pool,
	databaseCredentials *CmdDBCredential,
	appID CmdAppID,
) *Reindexer {
	panic(wire.Build(DependencySet))
}
