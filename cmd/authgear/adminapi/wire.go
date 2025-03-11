//go:build wireinject
// +build wireinject

package adminapi

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func NewInvoker(
	pool *db.Pool,
	credentials *config.GlobalDatabaseCredentialsEnvironmentConfig,
) *Invoker {
	panic(wire.Build(DependencySet))
}
