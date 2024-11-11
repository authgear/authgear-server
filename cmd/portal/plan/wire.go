//go:build wireinject
// +build wireinject

package plan

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func NewService(
	pool *db.Pool,
	databaseCredentials *config.DatabaseCredentials,
) *Service {
	panic(wire.Build(DependencySet))
}
